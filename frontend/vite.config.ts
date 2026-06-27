import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: 'inline',
      workbox: {
        cleanupOutdatedCaches: true,
        navigateFallbackDenylist: [/^\/api/],
      },
      manifest: {
        id: '/',
        name: 'Rosso Reader',
        short_name: 'Rosso',
        description: 'A self-hosted RSS feed reader',
        theme_color: '#f26522',
        background_color: '#ffffff',
        display: 'standalone',
        display_override: ['window-controls-overlay', 'standalone'],
        orientation: 'portrait',
        start_url: '/',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: 'maskable-icon-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
        screenshots: [
          {
            src: 'screenshot-desktop.png',
            sizes: '1024x811',
            type: 'image/png',
            form_factor: 'wide',
          },
          {
            src: 'screenshot-mobile.png',
            sizes: '563x1024',
            type: 'image/png',
          },
        ],
        protocol_handlers: [
          {
            protocol: 'web+feed',
            url: '/?feed=%s',
          },
        ],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
