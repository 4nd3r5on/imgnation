import { defineConfig } from 'vite'
import deno from '@deno/vite-plugin'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  server: {
    host: '0.0.0.0',
    port: 3000, // <--- Add this line for production port
  },
  dev: {
    host: '0.0.0.0',
  },
  plugins: [deno(), react(), tailwindcss()],
})
