import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: true,
    port: 5173,
    // El CORS del backend solo admite el puerto 5173: si está ocupado, falla en vez de saltar a otro.
    strictPort: true,
  },
})
