# Progressive Web App (PWA) Support

## Overview

Add Progressive Web App (PWA) support to Rosso Reader, allowing users to install the application on mobile and desktop devices. This includes offline caching of frontend assets, automatic updates in the background, a customized web app manifest, and proper server-side caching policies for the service worker.

## Decisions

- **Plugin:** `vite-plugin-pwa` for automated service worker generation and script injection.
- **Service Worker Mode:** `generateSW` with `registerType: 'autoUpdate'` for automatic background updates.
- **API Exclusions:** `navigateFallbackDenylist` set to `/^\/api/` so network requests to backend Go API routes bypass client-side SPA routing.
- **Nginx Headers:** Explicit cache bypass (`Cache-Control: no-store, no-cache`) for `/sw.js` and `.webmanifest` in the Nginx config to ensure clients fetch code updates immediately.

## Phase 1 — Asset Generation

Compile PWA-compliant PNG icons from the original `favicon.svg` using the system tool `rsvg-convert`:
- `pwa-192x192.png` & `pwa-512x512.png`: Standard square icons (with rounded corners) for desktop platforms.
- `maskable-icon-512x512.png`: Scaled by `0.75` and centered inside a solid `#f26522` viewport to accommodate safe-zone masking on Android.
- `apple-touch-icon.png`: Square icon (180x180) with a solid `#f26522` background for iOS Safari.

## Phase 2 — Frontend Configuration

### 2.1 Dependencies (`frontend/package.json`)
- Add `vite-plugin-pwa` to `devDependencies`.

### 2.2 Vite configuration (`frontend/vite.config.ts`)
- Load `VitePWA` in the plugins array:
```typescript
VitePWA({
  registerType: 'autoUpdate',
  injectRegister: 'inline',
  workbox: {
    cleanupOutdatedCaches: true,
    navigateFallbackDenylist: [/^\/api/],
  },
  manifest: {
    name: 'Rosso Reader',
    short_name: 'Rosso',
    description: 'A self-hosted RSS feed reader',
    theme_color: '#f26522',
    background_color: '#ffffff',
    display: 'standalone',
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
  },
})
```

### 2.3 Main entry (`frontend/index.html`)
- In `<head>`, add PWA tags:
```html
<link rel="apple-touch-icon" href="/apple-touch-icon.png" sizes="180x180" />
<meta name="theme-color" content="#f26522" />
```

## Phase 3 — Nginx Server Caching

### 3.1 Nginx config (`frontend/nginx.conf`)
- Configure `/sw.js` and `.webmanifest` locations to disable caching:
```nginx
    location = /sw.js {
        add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0";
        expires off;
    }

    location ~* \.webmanifest$ {
        add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0";
        expires off;
    }
```

## Phase 4 — Verification

### 4.1 Unit Tests
- Run `pnpm test` in the `frontend` folder to ensure all existing Vue/Pinia tests continue passing.

### 4.2 Production Build
- Run `pnpm build` in the `frontend` folder to verify:
  - Generation of `sw.js` and `workbox-*.js` files.
  - Manifest generation at `manifest.webmanifest`.
  - Successful injection of the inline service worker registration script and `<link rel="manifest">` tag into the built `index.html`.
