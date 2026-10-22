import { defineConfig } from '@vite-pwa/assets-generator/config'

// Genera los PNG de la PWA a partir de public/favicon.svg (que no es cuadrado: 48x46).
// Cada tamaño se dibuja centrado sobre un lienzo cuadrado con fondo de marca.
// Comando: npm run generate-pwa-assets
const background = '#1a1033'

export default defineConfig({
  headLinkOptions: { preset: '2023' },
  preset: {
    transparent: {
      sizes: [192, 512],
      favicons: [[48, 'favicon.ico']],
      padding: 0.1,
    },
    maskable: {
      sizes: [512],
      padding: 0.3,
      resizeOptions: { background, fit: 'contain' },
    },
    apple: {
      sizes: [180],
      padding: 0.3,
      resizeOptions: { background, fit: 'contain' },
    },
    assetName: (type, size) => {
      switch (type) {
        case 'transparent':
          return `icons/icon-${size.width}x${size.height}.png`
        case 'maskable':
          return `icons/icon-${size.width}x${size.height}-maskable.png`
        case 'apple':
          return `icons/apple-touch-icon-${size.width}x${size.height}.png`
      }
    },
    png: { compressionLevel: 9 },
  },
  images: ['public/favicon.svg'],
})
