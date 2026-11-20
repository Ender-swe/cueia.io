import type { components } from "@cueia/api-types"
export type Schema = components["schemas"]

export async function api<T>(path: string, token: string, method = "GET", body?: unknown, signal?: AbortSignal): Promise<T> {
  const headers: Record<string,string> = { Authorization: `Bearer ${token}` }
  if (body !== undefined) headers["Content-Type"] = "application/json"
  const res = await fetch("/api" + path, { method, headers, body: JSON.stringify(body), signal, cache: "no-store" })
  if (!res.ok) throw new Error(res.status === 503 ? "Workspace unavailable. Please retry." : "Request failed. Please retry.")
  return res.status === 204 ? (undefined as T) : await res.json()
}
