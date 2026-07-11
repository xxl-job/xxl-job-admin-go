import { defineConfig } from '@playwright/test'

// E2E 测试配置：默认指向本地 dev/preview (5173)，后端 REST 走 vite proxy 到 :18080。
// CI 通过 E2E_BASE_URL / E2E_SKIP_SERVER 环境变量覆盖。
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false, // 单实例跑，避免并发登录态干扰
  timeout: 30_000,
  retries: 1,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
    headless: true,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  // 本地复用已在跑的 dev server；CI 用 E2E_SKIP_SERVER 跳过自起（由 workflow 起 preview）
  webServer: process.env.E2E_SKIP_SERVER
    ? undefined
    : {
        command: 'pnpm dev --port 5173 --strictPort',
        url: 'http://localhost:5173',
        reuseExistingServer: true,
        timeout: 60_000,
      },
})
