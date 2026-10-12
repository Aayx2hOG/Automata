import { test, expect } from '@playwright/test';

test('Clerk sign-in and sign-up offer Google authentication', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: /Make the work/ })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Account' }).getByRole('button', { name: 'Sign in' })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Account' }).getByRole('button', { name: 'Sign up' })).toBeVisible();
  await expect(page.getByRole('button', { name: /Google/ }).first()).toBeVisible({ timeout: 30000 });
  await expect(page.getByRole('heading', { name: 'Welcome to Automata' })).toBeVisible();
  await expect(page.getByText('Loading Automata…', { exact: true })).toHaveCount(0);
  await page.screenshot({ path: '/tmp/automata-auth-desktop.png', fullPage: true });
  await page.getByRole('navigation', { name: 'Account' }).getByRole('button', { name: 'Sign up' }).click();
  await expect(page.getByRole('heading', { name: 'Create your account' })).toBeVisible();
  await expect(page.getByRole('button', { name: /Google/ }).first()).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByRole('button', { name: /Google/ }).first()).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: '/tmp/automata-auth-mobile.png', fullPage: true });
});
