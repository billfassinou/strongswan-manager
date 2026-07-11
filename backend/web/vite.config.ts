import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Le build est servi par le backend Go (même origine). En dev, on proxifie l'API et le
// WebSocket vers le backend local pour éviter tout souci CORS.
export default defineConfig({
  plugins: [react()],
  build: { outDir: "dist", emptyOutDir: true },
  server: {
    proxy: {
      "/api": { target: "http://localhost:8080", changeOrigin: true, ws: true },
      "/healthz": "http://localhost:8080",
    },
  },
});
