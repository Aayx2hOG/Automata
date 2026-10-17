import { expect, test } from '@playwright/test';

const workflows = [
  { id: 'alpha', name: 'Customer onboarding', description: 'Welcome new customers and notify the team.', is_active: true, updated_at: '2026-10-15T10:00:00Z', created_at: '2026-10-01T10:00:00Z' },
  { id: 'beta', name: 'Daily reporting', description: 'Collect metrics and deliver a daily summary.', is_active: false, updated_at: '2026-10-16T10:00:00Z', created_at: '2026-10-01T10:00:00Z' },
];
test.beforeEach(async ({ page }) => {
  await page.route('**/src/main.tsx', route => route.fulfill({ contentType: 'application/javascript', body: 'import "/tests/fixtures/navigation.tsx";' }));
  await page.route('**/api/**', route => {
    const path = new URL(route.request().url()).pathname;
    return route.fulfill({ json: path.endsWith('/workflows') ? { workflows } : path.endsWith('/schedules') ? { schedules: [] } : { runs: [] } });
  });
});

test('sections preserve browser history and keyboard search supports navigation', async ({ page }) => {
  await page.goto('/#/workflows');
  await expect(page.getByRole('heading', { name: 'Workflows', exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Executions', exact: true }).click();
  await expect(page).toHaveURL(/#\/runs$/);
  await page.goBack();
  await expect(page.getByRole('heading', { name: 'Workflows', exact: true })).toBeVisible();
  await page.keyboard.press('Control+k');
  await page.getByRole('combobox').fill('schedules');
  await page.keyboard.press('Enter');
  await expect(page).toHaveURL(/#\/schedules$/);
  await expect(page.getByRole('dialog')).toHaveCount(0);
});

test('workflow filters, view controls, and contextual links work', async ({ page }) => {
  await page.goto('/#/workflows');
  await page.getByRole('combobox', { name: 'Workflow status' }).selectOption('inactive');
  await expect(page.getByRole('button', { name: 'Daily reporting', exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Customer onboarding', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'List view', exact: true }).click();
  await expect(page.getByRole('button', { name: 'List view', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await page.getByTitle('Cron Schedule').click();
  await expect(page.locator('select').first()).toHaveValue('beta');
});

test('mobile navigation and dialogs remain usable with keyboard and small screens', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/');
  await page.getByRole('button', { name: 'Open navigation' }).click();
  await page.getByRole('link', { name: 'Workflows', exact: true }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await page.getByRole('button', { name: 'Add Workflow' }).click();
  const dialog = page.getByRole('dialog', { name: 'Create New Workflow' });
  await expect(dialog).toBeVisible();
  for (let index = 0; index < 12; index++) {
    await page.keyboard.press('Tab');
    expect(await dialog.evaluate(element => element.contains(document.activeElement))).toBe(true);
  }
  await page.keyboard.press('Escape');
  await expect(dialog).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Add Workflow' })).toBeFocused();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});

test('deletion requires confirmation and preserves the workflow on failure', async ({ page }) => {
  let deletions = 0;
  await page.route('**/api/workflows/alpha', route => {
    deletions++;
    return deletions === 1
      ? route.fulfill({ status: 409, json: { error: 'Workflow has queued or running executions.' } })
      : route.fulfill({ status: 204 });
  });
  await page.goto('/#/workflows');
  const trash = page.getByRole('button', { name: 'Delete Customer onboarding', exact: true });
  await trash.click();
  const dialog = page.getByRole('dialog', { name: 'Delete workflow?' });
  await expect(dialog).toContainText('Customer onboarding');
  await expect(dialog.getByRole('button', { name: 'Cancel' })).toBeFocused();
  await dialog.getByRole('button', { name: 'Cancel' }).click();
  expect(deletions).toBe(0);
  await expect(trash).toBeFocused();
  await trash.click();
  await dialog.getByRole('button', { name: 'Delete workflow', exact: true }).click();
  await expect(dialog.getByRole('alert')).toContainText('queued or running');
  await dialog.getByRole('button', { name: 'Delete workflow', exact: true }).click();
  await expect(dialog).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Customer onboarding', exact: true })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Daily reporting', exact: true })).toBeVisible();
  await expect(page.getByRole('searchbox', { name: 'Filter workflows' })).toBeFocused();
  // Stale polling responses must not resurrect a workflow after successful deletion.
  await expect(page.getByText('1 of 1 workflows', { exact: true })).toBeVisible();
});

test('an outdated backend produces a visible delete error instead of a silent failure', async ({ page }) => {
  await page.route('**/api/workflows/alpha', route => route.fulfill({ status: 405, body: '' }));
  await page.goto('/#/workflows');
  await page.getByRole('button', { name: 'Delete Customer onboarding', exact: true }).click();
  const dialog = page.getByRole('dialog', { name: 'Delete workflow?' });
  await dialog.getByRole('button', { name: 'Delete workflow', exact: true }).click();
  await expect(dialog.getByRole('alert')).toContainText('Restart the backend');
  await expect(dialog.getByRole('button', { name: 'Delete workflow', exact: true })).toBeEnabled();
  await dialog.getByRole('button', { name: 'Cancel' }).click();
  await expect(page.getByRole('button', { name: 'Customer onboarding', exact: true })).toBeVisible();
});
