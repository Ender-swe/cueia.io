import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig, loadEnv } from "vite"

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "CUEIA_")
  const proxy = { "/api": { target: env.CUEIA_API_ORIGIN || "http://127.0.0.1:8080" } }
  return { plugins: [react(), tailwindcss()], server: { host: "127.0.0.1", proxy }, preview: { host: "127.0.0.1", proxy } }
})
