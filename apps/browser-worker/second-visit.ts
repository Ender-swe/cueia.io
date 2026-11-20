import { signInTestAccount } from "./test-login"
import { chromium, expect } from "@playwright/test"
import { createHash } from "node:crypto"
import { readFile, writeFile } from "node:fs/promises"
import { join } from "node:path"

const origin = process.env.CLERK_AUTHORIZED_PARTY!
const state = process.env.CUEIA_ACCEPTANCE_DIR!
const disk = process.env.CUEIA_TEST_DISK_DIR!
const project = process.env.COMPOSE_PROJECT_NAME!
if (!origin.startsWith("http://127.0.0.1:") || !state || !project || !disk) throw Error("Missing local acceptance configuration")
const alice = JSON.parse(await readFile(join(state,"alice.json"),"utf8"))
const email = process.env.CUEIA_TEST_BOB_EMAIL?.trim()
if (!email || !process.env.CLERK_SECRET_KEY?.startsWith("sk_test_")) throw Error("Configure the second Clerk test identity")
const browser = await chromium.launch()
try {
  const page=await browser.newPage()
  await page.goto(origin)
  await signInTestAccount(page,email)
  const exchange=page.waitForResponse(r=>r.url()===origin+"/api/session" && r.ok(),{timeout:300000})
  await page.getByRole("link",{name:"Open sample workspace"}).click()
  const {token}=await (await exchange).json()
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible()
  const request=(bearer:string,path:string,method="GET")=>fetch(origin+"/api"+path,{method,headers:{Authorization:"Bearer "+bearer}})
  const res=await request(token,"/workspace")
  expect(res.status).toBe(200)
  const snapshot=await res.json()
  expect(snapshot.id).not.toBe(alice.snapshot.id)
  expect(snapshot.preferences.timeFormat).toBe("12h")
  expect(snapshot.threads[0].id).not.toBe(alice.snapshot.threads[0].id)
  const attachment=snapshot.threads[0].messages[0].attachments[0]
  expect(attachment.fileId).not.toBe(alice.attachment.fileId)
  const hash=createHash("sha256").update(JSON.stringify({Deployment:project,Workspace:snapshot.id})).digest("hex")
  const root=join(disk,hash,"disk","workspace")
  const name="owner-notes.csv"
  for(const [dir,bearer,value] of [[alice.root,alice.sessions[1],"Alice"],[root,token,"Bob"]]){
    await writeFile(join(dir,name),value,{flag:"w"})
    const entries=await (await request(bearer,"/files")).json()
    expect(entries.some((e:{name:string})=>e.name===name)).toBe(true)
    expect(await (await request(bearer,"/files/content?path="+name)).text()).toBe(value)
  }
  expect((await request(token,"/documents/"+alice.attachment.fileId+"/content")).status).toBe(404)
  expect((await request(token,"/documents/"+alice.attachment.fileId+"/preview-links","POST")).status).toBe(404)
  const bytes=Buffer.from(await (await request(token,"/documents/"+attachment.fileId+"/content")).arrayBuffer())
  expect(bytes.toString("base64")===alice.pdf).toBe(true)
  const link=await (await request(token,"/documents/"+attachment.fileId+"/preview-links","POST")).json()
  const preview=await fetch(origin+link.url)
  expect(preview.ok && Buffer.from(await preview.arrayBuffer()).equals(bytes)).toBe(true)
  await writeFile(join(state,"bob.json"),JSON.stringify({sessions:[token],snapshot,attachment,pdf:bytes.toString("base64"),root}),{mode:0o600,flag:"wx"})
  console.log("Bob's GitHub sign-in and owner isolation checks passed.")
} finally {
  await browser.close()
}
