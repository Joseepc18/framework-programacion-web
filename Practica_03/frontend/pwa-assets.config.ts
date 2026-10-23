import { defineConfig } from '@vite-pwa/assets-generator/config'

// Genera los PNG de la PWA a partir del SVG cuadrado de 512 px.
// El ícono de tienda queda dentro del círculo seguro del icono maskable.
// Comando: npm run generate-pwa-assets
const background = '#0e2a47'

export default defineConfig({
  headLinkOptions: { preset: '2023' },
  preset: {
    transparent: {
      sizes: [192, 512],
      favicons: [[48, 'favicon.ico']],
      padding: 0,
    },
    maskable: {
      sizes: [512],
      padding: 0,
      resizeOptions: { background, fit: 'contain' },
    },
    apple: {
      sizes: [180],
      padding: 0,
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
