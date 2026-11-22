import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { fileURLToPath } from "node:url";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  const envDir = fileURLToPath(new URL("../../", import.meta.url));
  const env = loadEnv(mode, envDir, "CUEIA_");
  const proxy = {
    "/api": { target: env.CUEIA_API_ORIGIN || "http://127.0.0.1:8080" },
  };
  return {
    envDir,
    plugins: [react(), tailwindcss()],
    server: { host: "127.0.0.1", proxy },
    preview: { host: "127.0.0.1", proxy },
  };
});
