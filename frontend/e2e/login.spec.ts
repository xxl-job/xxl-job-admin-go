import { test, expect } from '@playwright/test'

// 登录 E2E：验证前端 → REST → 后端 → MySQL 全链路。
// 前置：后端 admin+mysql 已起（docker compose），admin/123456 已 seed。
test('登录成功进入仪表盘', async ({ page }) => {
  await page.goto('/')
  // 登录页（未登录重定向到 /login）
  await expect(page).toHaveURL(/\/login/)

  await page.getByRole('textbox', { name: '用户名' }).fill('admin')
  await page.getByRole('textbox', { name: '密码' }).fill('123456')
  await page.getByRole('button', { name: '登录' }).click()

  // 登录成功跳转首页（仪表盘）
  await expect(page).toHaveURL(/\/$|^\S+\/$/)
  await expect(page.getByText('调度仪表盘')).toBeVisible({ timeout: 10_000 })
})

test('错误密码登录失败', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('textbox', { name: '用户名' }).fill('admin')
  await page.getByRole('textbox', { name: '密码' }).fill('wrong-password')
  await page.getByRole('button', { name: '登录' }).click()

  // 应停留在登录页并提示错误（ElMessage 文案）
  await expect(page).toHaveURL(/\/login/)
  await expect(page.getByText(/用户名或密码错误|登录失败/)).toBeVisible({ timeout: 10_000 })
})
