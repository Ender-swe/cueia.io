import { chromium, expect } from "@playwright/test"
import { createHash, randomUUID } from "node:crypto"
import { mkdir, readFile, writeFile } from "node:fs/promises"
import { join } from "node:path"
import expected from "./fixtures/north-pier-expected.json"

const origin = process.env.CUEIA_DASHBOARD_ORIGIN ?? "http://127.0.0.1:5173"
const disk = process.env.CUEIA_TEST_DISKS
const deployment = process.env.CUEIA_TEST_DEPLOYMENT
if (!disk || !deployment) throw Error("Run the dashboard journey with the internal workspace provider")

const browser = await chromium.launch({headless:false})
try {
  const page = await browser.newPage({viewport:{width:1280,height:840}, locale:"en-US"})
  page.on("response", response => {
    if (response.url() === origin + "/api/session" && !response.ok()) {
      console.error("Workspace session exchange returned", response.status())
    }
  })
  await page.goto(origin)
  const open = page.getByRole("link", {name:"Open sample workspace"})
  const snapshot = page.waitForResponse(r => r.url() === origin + "/api/workspace" && r.ok(), {timeout:300000})
  await open.click()
  console.log("Complete sign-in in the Chromium window.")
  const {id} = await (await snapshot).json()
  const workspaceHash = createHash("sha256").update(JSON.stringify({Deployment:deployment, Workspace:id})).digest("hex")
  const root = join(disk, workspaceHash, "disk", "workspace")
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible()
  // Returning from the launch page must reuse the current Clerk session.
  await page.goto(origin)
  await open.click()
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible({timeout:15000})
  await expect(page.getByText(expected.invitationSender)).toBeVisible()
  const event = page.locator("time")
  const date = new Date((await event.getAttribute("datetime"))!)
  expect(date.toLocaleDateString("sv-SE",{timeZone:expected.workshopTimeZone})).toBe(expected.workshopLocalDate)
  expect(date.toLocaleTimeString("en-GB",{timeZone:expected.workshopTimeZone,hour:"2-digit",minute:"2-digit"})).toBe(expected.workshopLocalTime)
  await expect(event).toContainText(expected.workshopLocalTime)
  const draft = page.getByRole("textbox",{name:"Message",exact:true})
  await draft.fill("Bring the field notes.")
  const button = (name:string) => page.getByRole("button",{name,exact:true})
  await button("Tasks").click()
  const tasks = page.getByRole("complementary",{name:"Tasks"})
  await expect(tasks.getByText("Agent disconnected",{exact:false})).toBeVisible()
  for(const label of ["Scheduled tasks","No scheduled tasks.","Today's activity","No activity yet."]) {
    await expect(tasks.getByText(label,{exact:true})).toBeVisible()
  }
  await button("Files").click()
  const folder = "field-notes-" + randomUUID()
  await mkdir(join(root,folder))
  await writeFile(join(root,folder,"attendance.csv"),"name,attending\nMara,yes\n",{flag:"wx"})
  await button("Refresh").click()
  await button(folder+"/").click()
  await expect(page.getByText("attendance.csv",{exact:true})).toBeVisible()
  const search = page.getByRole("searchbox",{name:"Search files"})
  await search.fill("attendance.csv")
  await expect(page.getByText(folder+"/attendance.csv",{exact:true})).toBeVisible()
  await search.fill("missing-"+folder)
  await expect(page.getByText("No matching files.")).toBeVisible()
  await search.fill(expected.attachmentName)
  const file = button(expected.attachmentName)
  const preview = page.waitForResponse(r => r.url().startsWith(origin+"/api/previews/") && r.ok())
  await file.click()
  const pdf = await readFile(new URL("../../backend/internal/workspace/sample/north-pier-registration.pdf",import.meta.url))
  await preview
  const frame = page.getByTitle(expected.attachmentName, {exact:true})
  await expect(frame).toBeVisible()
  const pdfPath = await frame.getAttribute("src")
  if (!pdfPath?.startsWith("/api/previews/")) throw Error("Preview frame did not reference the signed document")
  const response = await page.request.get(origin + pdfPath)
  expect(response.ok() && (await response.body()).equals(pdf)).toBe(true)
  await expect(page.getByRole("dialog")).toBeVisible()
  await page.keyboard.press("Escape")
  await expect(file).toBeFocused()
  await expect(tasks).toBeVisible()
  await button("Chat").click()
  await button("Close Tasks").click()
  await expect(draft).toHaveValue("Bring the field notes.")
  await page.setViewportSize({width:390,height:844})
  await button("Files").click()
  await button("Close sidebar").click()
  await button("Tasks").click()
  await button("Close Tasks").click()
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible()
  await draft.fill("Ready for Saturday.")
  await expect(button("Call")).toHaveCount(0)
  await expect(button("Invite")).toHaveCount(0)
  console.log("Dashboard journey passed.")
} finally {
  await browser.close()
}
