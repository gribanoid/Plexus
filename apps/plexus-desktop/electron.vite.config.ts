import { resolve } from 'path'
import { defineConfig, externalizeDepsPlugin } from 'electron-vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  main: {
    plugins: [externalizeDepsPlugin()],
    resolve: {
      alias: {
        '@main': resolve('src/main'),
      },
    },
  },
  preload: {
    plugins: [externalizeDepsPlugin()],
  },
  renderer: {
    resolve: {
      alias: {
        '@': resolve('src/renderer'),
        '@plexus/api': resolve('../../packages/api/src'),
        '@plexus/ui': resolve('../../packages/ui/src'),
        '@plexus/features': resolve('../../packages/features/src'),
      },
    },
    plugins: [react(), tailwindcss()],
  },
})
