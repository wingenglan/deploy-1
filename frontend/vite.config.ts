import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// defineConfig keeps local API requests on the same /api contract as production.
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8081',
    },
  },
})
