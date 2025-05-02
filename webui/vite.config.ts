import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  css: {
    postcss: "./postcss.config.cjs",
  },
  server: {
    port: 6936,
    open: true,
    host: true,
    proxy: {
      "/apis": {
        target: "http://localhost:6937",
        changeOrigin: true,
      },
    },
  },
});
