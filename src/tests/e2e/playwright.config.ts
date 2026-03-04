import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  use: {
    baseURL: 'http://127.0.0.1:5173'
  }
})
