import { chromium, expect } from "@playwright/test"
import { createHash } from "node:crypto"
import { join } from "node:path"
import { signInTestAccount } from "./test-login"
import { checkDashboard } from "./check-dashboard"

const origin = process.env.CUEIA_DASHBOARD_ORIGIN ?? "http://127.0.0.1:5173"
const disk = process.env.CUEIA_TEST_DISKS
const deployment = process.env.CUEIA_TEST_DEPLOYMENT
if (!disk || !deployment) throw Error("Run the dashboard journey with the internal workspace provider")

const email = process.env.CUEIA_TEST_USER_EMAIL?.trim()
if (email && !process.env.CLERK_SECRET_KEY?.startsWith("sk_test_")) {
  throw Error("Automated Clerk login requires the test instance's CLERK_SECRET_KEY")
}
const browser = await chromium.launch({headless: Boolean(email)})
try {
  const page = await browser.newPage({viewport:{width:1280,height:840}, locale:"en-US"})
  page.on("response", response => {
    if (response.url() === origin + "/api/session" && !response.ok()) {
      console.error("Workspace session exchange returned", response.status())
    }
  })
  await page.goto(origin)
  if (email) {
    await signInTestAccount(page, email)
  }
  const open = page.getByRole("link", {name:"Open sample workspace"})
  const snapshot = page.waitForResponse(r => r.url() === origin + "/api/workspace" && r.ok(), {timeout:300000})
  await open.click()
  if (!email) console.log("Complete sign-in in the Chromium window.")
  const {id} = await (await snapshot).json()
  const workspaceHash = createHash("sha256").update(JSON.stringify({Deployment:deployment, Workspace:id})).digest("hex")
  const root = join(disk, workspaceHash, "disk", "workspace")
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible()
  // Returning from the launch page must reuse the current Clerk session.
  await page.goto(origin)
  await open.click()
  await expect(page.getByRole("heading",{name:"Main chat"})).toBeVisible({timeout:15000})
  await checkDashboard(page, origin, root)
  console.log("Dashboard journey passed.")
} finally {
  await browser.close()
}
