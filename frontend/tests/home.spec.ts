import { expect, test } from '@playwright/test';

test('empty workspace guides users to keyboard-accessible templates', async ({ page }) => {
  await page.route('**/src/main.tsx', route => route.fulfill({ contentType: 'application/javascript', body: 'import "/tests/fixtures/home.tsx";' }));
  await page.goto('/');
  await expect(page.getByText('Your first workflow starts here')).toBeVisible();
  await page.getByRole('button', { name: 'Choose a template' }).click();
  const template = page.getByRole('button', { name: /Conditional Status Router/ });
  await template.focus();
  await page.keyboard.press('Enter');
  await expect(template).toHaveAttribute('aria-pressed', 'true');
  await page.getByRole('button', { name: 'Close new workflow' }).click();
  await page.getByRole('button', { name: 'View execution history' }).click();
  await expect(page.locator('body')).toHaveAttribute('data-tab', 'runs');
  await page.setViewportSize({ width: 390, height: 844 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});
