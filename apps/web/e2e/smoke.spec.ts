import { test, expect } from '@playwright/test'

test.describe('Plexus smoke', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: /log in to continue/i })).toBeVisible()
    await expect(page.getByLabel(/email or username/i)).toBeVisible()
    await expect(page.getByLabel(/^password$/i)).toBeVisible()
    await expect(page.getByRole('button', { name: /continue with sso/i })).toBeVisible()
  })

  test('unauthenticated board redirect', async ({ page }) => {
    await page.goto('/orgs/plexus-dev/DEMO/board')
    await expect(page).toHaveURL(/\/login/)
  })

  test('auth callback page handles missing tokens', async ({ page }) => {
    await page.goto('/auth/callback')
    await expect(page).toHaveURL(/\/login/)
  })
})
