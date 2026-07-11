import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Le build est servi par le backend Go (même origine). En dev, on proxifie l'API et le
// WebSocket vers le backend local pour éviter tout souci CORS.
//
// Le backend sert en HTTPS avec un certificat auto-signé : `secure: false` demande au proxy
// de ne pas en vérifier la chaîne — sans quoi `npm run dev` échouerait sur un certificat que
// Node ne connaît pas. Cela ne concerne QUE le serveur de développement local.
const backend = "https://localhost:7926";

export default defineConfig({
  plugins: [react()],
  build: { outDir: "dist", emptyOutDir: true },
  server: {
    proxy: {
      "/api": { target: backend, changeOrigin: true, ws: true, secure: false },
      "/healthz": { target: backend, changeOrigin: true, secure: false },
    },
  },
});
