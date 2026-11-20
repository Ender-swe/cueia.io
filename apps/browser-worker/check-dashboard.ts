import { expect, type Page } from "@playwright/test"
import { randomUUID } from "node:crypto"
import { mkdir, readFile, writeFile } from "node:fs/promises"
import { join } from "node:path"
import expected from "./fixtures/north-pier-expected.json"

async function checkResize(page: Page, name: string, increase: number) {
  const divider = page.getByRole("separator", {name, exact:true})
  const panel = page.getByRole("complementary", {name, exact:true})
  const nav = page.getByRole("navigation", {name:"Workspace"})
  const before = await panel.boundingBox()
  const handle = await divider.boundingBox()
  const rail = await nav.boundingBox()
  if (!before || !handle || !rail) throw Error("Missing resize controls for " + name)
  const stacked = await divider.getAttribute("aria-orientation") === "horizontal"
  const x = handle.x + handle.width / 2
  const y = handle.y + handle.height / 2
  const dx = name === "Tasks" ? -increase : increase
  await page.mouse.move(x, y)
  await page.mouse.down()
  await page.mouse.move(stacked ? x : x + dx, stacked ? y + increase : y, {steps:4})
  await page.mouse.up()
  const dimension = stacked ? "height" : "width"
  const size = () => panel.evaluate((el, key: "width" | "height") => el.getBoundingClientRect()[key], dimension)
  await expect.poll(size).toBe(before[dimension] + increase)
  await divider.press(stacked ? "ArrowUp" : name === "Tasks" ? "ArrowRight" : "ArrowLeft")
  await expect.poll(size).toBe(before[dimension] + increase - 16)
  expect(await nav.boundingBox()).toEqual(rail)
}

export async function checkDashboard(page: Page, origin: string, root: string) {
  await expect(page.getByText(expected.invitationSender)).toBeVisible()
  const event = page.locator("time")
  const date = new Date((await event.getAttribute("datetime"))!)
  expect(date.toLocaleDateString("sv-SE",{timeZone:expected.workshopTimeZone})).toBe(expected.workshopLocalDate)
  expect(date.toLocaleTimeString("en-GB",{timeZone:expected.workshopTimeZone,hour:"2-digit",minute:"2-digit"})).toBe(expected.workshopLocalTime)
  await expect(event).toContainText(expected.workshopLocalTime)
  const draft = page.getByRole("textbox",{name:"Message",exact:true})
  await draft.fill("Bring the field notes.")
  const button = (name:string) => page.getByRole("button",{name,exact:true})
  await button("Menu").click()
  const preference = page.getByLabel("Time format")
  const originalFormat = await preference.inputValue()
  const otherFormat = originalFormat === "12h" ? "24h" : "12h"
  for (const value of [otherFormat, originalFormat]) {
    const persisted = page.waitForResponse(response =>
      response.url() === origin + "/api/workspace/preferences" && response.request().method() === "PATCH" && response.ok()
    )
    await preference.selectOption(value)
    await persisted
    await expect(preference).toHaveValue(value)
    if (value === "12h") await expect(event).toContainText("AM")
    else await expect(event).not.toContainText("AM")
  }
  await page.keyboard.press("Escape")
  await expect(button("Menu")).toBeFocused()
  const nav = page.getByRole("navigation", {name:"Workspace"})
  for (const width of [1280, 390]) {
    await page.setViewportSize({width, height:840})
    const content = await page.locator(".conversation").boundingBox()
    if (!content) throw Error("Missing chat view")
    let bottom = 0
    for (const name of ["Chat", "Files", "Tasks"]) {
      const control = nav.getByRole("button", {name, exact:true})
      await expect(control).toBeInViewport()
      const box = await control.boundingBox()
      if (!box) throw Error("Missing sidebar button: " + name)
      expect(box.x + box.width).toBeLessThanOrEqual(content.x)
      expect(box.y).toBeGreaterThanOrEqual(bottom)
      bottom = box.y + box.height
    }
  }
  await page.setViewportSize({width:1280,height:840})
  await button("Tasks").click()
  const tasks = page.getByRole("complementary",{name:"Tasks"})
  await expect(tasks.getByText("Agent disconnected",{exact:false})).toBeVisible()
  for(const label of ["Scheduled tasks","No scheduled tasks.","Today's activity","No activity yet."]) {
    await expect(tasks.getByText(label,{exact:true})).toBeVisible()
  }
  await button("Files").click()
  await expect(tasks).toBeVisible()
  const files = page.getByRole("complementary",{name:"Files"})
  const conversation = page.getByRole("region",{name:"Main chat"})
  function bounds(element: typeof files) {
    return element.evaluate(el => ({left:el.getBoundingClientRect().left, right:el.getBoundingClientRect().right}))
  }
  const [railBox, filesBox, chatBox, tasksBox] = await Promise.all([nav, files, conversation, tasks].map(bounds))
  expect(railBox.right).toBeLessThanOrEqual(filesBox.left + 1)
  expect(filesBox.right).toBeLessThanOrEqual(chatBox.left + 1)
  expect(chatBox.right).toBeLessThanOrEqual(tasksBox.left + 1)
  await checkResize(page, "Files", 48)
  expect(await bounds(tasks)).toEqual(tasksBox)
  const resizedFiles = await bounds(files)
  await checkResize(page, "Tasks", 48)
  expect(await bounds(files)).toEqual(resizedFiles)
  await expect(draft).toBeInViewport({ratio:1})
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
  const pdf = await readFile(new URL("../server/internal/workspace/sample/north-pier-registration.pdf",import.meta.url))
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
  await button("Chat").click()
  await expect(page.getByRole("searchbox",{name:"Search chats"})).toBeVisible()
  expect(await bounds(page.getByRole("complementary",{name:"Chats"}))).toEqual(resizedFiles)
  await expect(tasks).toBeVisible()
  await button("Close Tasks").click()
  await expect(button("Tasks")).toBeFocused()
  await expect(tasks).toBeHidden()
  await page.setViewportSize({width:390,height:844})
  await page.emulateMedia({reducedMotion: "reduce"})
  await button("Files").click()
  await expect(search).toHaveValue(expected.attachmentName)
  await checkResize(page, "Files", 32)
  await expect(draft).toBeInViewport({ratio:1})
  await button("Tasks").click()
  await expect(tasks).toBeInViewport()
  await checkResize(page, "Tasks", 24)
  expect((await bounds(conversation)).right).toBeLessThanOrEqual((await bounds(tasks)).left + 1)
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeInViewport()
  await expect(draft).toBeInViewport({ratio:1})
  await expect(draft).toHaveValue("Bring the field notes.")
  await page.setViewportSize({width:390,height:460})
  await draft.fill("Ready for Saturday.")
  await expect(draft).toBeInViewport({ratio:1})
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await expect(button("Call")).toHaveCount(0)
  await expect(button("Invite")).toHaveCount(0)
}
