import type { Page } from "@playwright/test"
import { readFile } from "node:fs/promises"
import { parseEnv } from "node:util"

export async function signInTestAccount(page: Page, email: string) {
  const { clerk, clerkSetup } = await import("@clerk/testing/playwright")
  const frontendApiUrl = new URL(process.env.CUEIA_BROWSER_CLERK_ISSUER!).hostname
  const publishableKey = process.env.VITE_CLERK_PUBLISHABLE_KEY || process.env.CLERK_PUBLISHABLE_KEY ||
    parseEnv(await readFile(new URL("../dashboard/.env.local", import.meta.url), "utf8")).VITE_CLERK_PUBLISHABLE_KEY
  if (!publishableKey?.startsWith("pk_test_")) throw Error("Expected the dashboard's Clerk test publishable key")
  await clerkSetup({ frontendApiUrl, publishableKey, dotenv: false })
  await clerk.signIn({ page, emailAddress: email })
}
