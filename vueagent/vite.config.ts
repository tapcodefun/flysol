import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  server: {
    port: 3000, // 设置默认端口
  },
  plugins: [vue()],
})
