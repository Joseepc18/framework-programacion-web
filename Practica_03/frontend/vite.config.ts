import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      // 'prompt': el SW nuevo espera a que el usuario pulse "Actualizar" (ReloadPrompt).
      registerType: 'prompt',
      includeAssets: ['favicon.svg', 'favicon.ico', 'icons/apple-touch-icon-180x180.png'],
      manifest: {
        name: 'MultiCatálogo',
        short_name: 'MultiCat',
        description: 'Tienda cosmética con red MLM multinivel',
        theme_color: '#1a1033',
        background_color: '#1a1033',
        display: 'standalone',
        start_url: '/',
        scope: '/',
        icons: [
          { src: '/icons/icon-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512x512.png', sizes: '512x512', type: 'image/png' },
          { src: '/icons/icon-512x512-maskable.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      workbox: {
        // Precache de assets estáticos del build.
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        // Las rutas del router (/tienda, /producto/:id) abren offline con el index.html.
        navigateFallback: '/index.html',
        runtimeCaching: [
          {
            // La API vive en otro origen (localhost:3001): se compara contra url.pathname.
            // Solo catálogo público; NUNCA /api/red, checkout, pedidos ni liquidaciones.
            urlPattern: ({ url }) => url.pathname.startsWith('/api/productos'),
            handler: 'NetworkFirst',
            method: 'GET',
            options: {
              // Debe coincidir con API_CACHE_NAME en src/services/pwaCache.ts
              cacheName: 'api-productos',
              networkTimeoutSeconds: 5,
              expiration: { maxEntries: 50, maxAgeSeconds: 60 * 60 * 24 },
              cacheableResponse: { statuses: [200] },
            },
          },
        ],
      },
    }),
  ],
  server: {
    host: true,
  },
})
