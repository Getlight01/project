import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3005,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8085',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8085',
        changeOrigin: true,
      },
      '/looks': {
        target: 'http://localhost:8085',
        changeOrigin: true,
      },
      '/models': {
        target: 'http://localhost:8085',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
