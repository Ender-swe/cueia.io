import { chromium, expect } from "@playwright/test"
import { createHash } from "node:crypto"
import { readFile, readdir, writeFile } from "node:fs/promises"
import { join } from "node:path"
import { signInTestAccount } from "./test-login"
import { checkDashboard } from "./check-dashboard"

const origin = process.env.CLERK_AUTHORIZED_PARTY!
const disks = process.env.CUEIA_TEST_DISK_DIR!
const project = process.env.COMPOSE_PROJECT_NAME!
const state = process.env.CUEIA_ACCEPTANCE_DIR!
if (!origin.startsWith("http://127.0.0.1:") || !state || !project || !disks) throw Error("Missing local acceptance configuration")
const email = process.env.CUEIA_TEST_USER_EMAIL?.trim()
if (!email || !process.env.CLERK_SECRET_KEY?.startsWith("sk_test_")) throw Error("Configure the Google test account before running the first visit")
if ((await readdir(disks)).length) throw Error("The first visit requires an empty disposable workspace provider")
const browser = await chromium.launch()
try {
  const page = await browser.newPage({viewport:{width:1280,height:840},locale:"en-US"})
  const sessions: string[] = []
  await page.route(origin+"/api/session",async route => {
    if (route.request().method() !== "POST" || sessions.length) return route.continue()
    const [a,b] = await Promise.all([route.fetch(),route.fetch()])
    if (!a.ok() || !b.ok()) throw Error("Concurrent exchanges failed")
    sessions.push((await a.json()).token,(await b.json()).token)
    await route.fulfill({response:a})
  })
  await page.goto(origin)
  await signInTestAccount(page,email)
  await page.getByRole("link",{name:"Open sample workspace"}).click()
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible({timeout:300000})
  if (sessions.length!==2 || !sessions[0] || sessions[0]===sessions[1]) throw Error("Expected two distinct sessions")
  async function api(token:string, path:string, method="GET", body?:unknown) {
    const response = await fetch(origin+"/api"+path,{method,headers:{Authorization:"Bearer "+token,"Content-Type":"application/json"},body:JSON.stringify(body)})
    if (!response.ok) throw Error("Workspace request failed with HTTP "+response.status)
    return response
  }
  const [a,b] = sessions
  const snapshot = (token:string) => api(token,"/workspace").then(r=>r.json())
  const first=await snapshot(a)
  expect(first).toEqual(await snapshot(b))
  expect(first.threads).toHaveLength(1)
  expect(first.events).toHaveLength(1)
  const attachment=first.threads[0].messages[0].attachments[0]
  const hash=createHash("sha256").update(JSON.stringify({Deployment:project,Workspace:first.id})).digest("hex")
  expect(await readdir(disks)).toEqual([hash])
  const root=join(disks,hash,"disk","workspace")
  await checkDashboard(page,origin,root)
  await api(a,"/workspace/preferences","PATCH",{timeFormat:"24h"})
  const saved=await snapshot(b)
  expect(saved.preferences.timeFormat).toBe("24h")
  const bytes=Buffer.from(await (await api(b,"/documents/"+attachment.fileId+"/content")).arrayBuffer())
  const source=await readFile(new URL("../server/internal/workspace/sample/north-pier-registration.pdf",import.meta.url))
  expect(bytes.equals(source)).toBe(true)
  const filename="workshop-copy.pdf"
  await writeFile(join(root,filename),source,{flag:"wx"})
  const register = (token:string) => api(token,"/documents","POST",{path:filename}).then(r=>r.json())
  const ids=await Promise.all([register(a),register(b)])
  expect(ids[0].fileId).toBe(ids[1].fileId)
  const link = (id:string) => api(a,"/documents/"+id+"/preview-links","POST").then(r=>r.json())
  const old=await link(ids[0].fileId)
  const edited=Buffer.concat([source,Buffer.from("\n% revised workshop copy\n")])
  await writeFile(join(root,filename),edited)
  const updated=await register(a)
  expect(updated.fileId).not.toBe(ids[0].fileId)
  for(const [url,expected] of [[old.url,source],[(await link(updated.fileId)).url,edited]] as const) {
    const response=await fetch(origin+url)
    expect(response.ok && Buffer.from(await response.arrayBuffer()).equals(expected)).toBe(true)
  }
  await writeFile(join(state,"alice.json"),JSON.stringify({sessions,snapshot:saved,attachment,root,pdf:bytes.toString("base64")}),{mode:0o600,flag:"wx"})
  console.log("Alice's first-visit browser and storage checks passed.")
} finally {
  await browser.close()
}
