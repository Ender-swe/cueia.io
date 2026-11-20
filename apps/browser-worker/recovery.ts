import { strict as assert } from "node:assert"
import { readFile, writeFile } from "node:fs/promises"
import { join } from "node:path"

const origin = process.env.CLERK_AUTHORIZED_PARTY!
const state = process.env.CUEIA_ACCEPTANCE_DIR!
if (!origin.startsWith("http://127.0.0.1:") || !state) throw Error("Missing local recovery settings")
const load = async (name:string) => JSON.parse(await readFile(join(state,name+".json"),"utf8"))
const alice = await load("alice")
const bob = await load("bob")
const path = "/api/documents/"+alice.attachment.fileId+"/preview-links"
async function request(path:string,token?:string,method="GET") {
  return fetch(origin+path,{method,headers:token ? {Authorization:"Bearer "+token} : {}})
}
async function issue(token:string) {
  const r=await request(path,token,"POST")
  assert.equal(r.status,200)
  return (await r.json()).url
}
async function checkPdf(url:string,pdf:string) {
  const r=await request(url)
  assert.equal(r.status,200)
  assert.ok(Buffer.from(await r.arrayBuffer()).toString("base64")===pdf,"PDF bytes differed")
}
const [a,b] = alice.sessions
const step=process.argv[2]
if (step==="capture") {
  await writeFile(join(state,"preview.json"),JSON.stringify(await issue(a)),{mode:0o600})
} else if (step==="restarted" || step==="replaced") {
  for (const [owner,label] of [[alice,"Alice"],[bob,"Bob"]]) {
    for (const token of owner.sessions) {
      const r=await request("/api/workspace",token)
      assert.equal(r.status,200)
      assert.deepEqual(await r.json(),owner.snapshot)
    }
    const token=owner.sessions[0]
    const doc=await request("/api/documents/"+owner.attachment.fileId+"/content",token)
    assert.equal(doc.status,200)
    assert.ok(Buffer.from(await doc.arrayBuffer()).toString("base64")===owner.pdf,"Attachment changed")
    const entries=await (await request("/api/files",token)).json()
    assert.ok(entries.some((e:{name:string})=>e.name==="owner-notes.csv"))
    assert.equal(await (await request("/api/files/content?path=owner-notes.csv",token)).text(),label)
  }
  await checkPdf(await load("preview"),alice.pdf)
  if (step==="replaced") {
    assert.equal((await request("/api/session",a,"DELETE")).status,204)
    assert.equal((await request("/api/workspace",a)).status,401)
    for (const [token,owner] of [[b,alice],[bob.sessions[0],bob]]) {
      const r=await request("/api/workspace",token)
      assert.equal(r.status,200)
      assert.equal((await r.json()).id,owner.snapshot.id)
    }
  }
} else if (step==="expired") {
  assert.equal((await request(await load("preview"))).status,403)
  await checkPdf(await issue(b),alice.pdf)
} else throw Error("Use capture, restarted, replaced, or expired")
console.log("Recovery check passed:",step)
