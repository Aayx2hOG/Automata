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
  await page.getByRole('searchbox', { name: 'Filter workflows' }).fill('Daily');
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


test('schedule picker sends custom time options to the backend', async ({ page }) => {
  await page.route('**/api/workflows/alpha/schedules', route => route.fulfill({ json: route.request().method() === 'POST' ? { id: 'schedule-new', workflow_id: 'alpha', is_active: true, cron_expression: 'CRON_TZ=Asia/Kolkata 45 14 * * 2' } : { schedules: [] } }));
  await page.goto('/#/schedules');
  await page.getByLabel('Repeat', { exact: true }).selectOption('weekly');
  await page.getByLabel('Day of week').selectOption('2');
  await page.getByLabel('Time', { exact: true }).fill('14:45');
  await page.getByText('Change time zone', { exact: true }).click();
  await page.getByLabel('Time zone', { exact: true }).selectOption('Asia/Kolkata');
  const request = page.waitForRequest(request => request.method() === 'POST' && request.url().endsWith('/alpha/schedules'));
  await page.getByRole('button', { name: 'Create schedule' }).click();
  expect((await request).postDataJSON()).toEqual({ options: { frequency: 'weekly', time: '14:45', weekday: 2, day: 1, minute: 45, timezone: 'Asia/Kolkata' } });
  await expect(page.getByText('CRON_TZ=Asia/Kolkata 45 14 * * 2', { exact: true })).toBeVisible();
});

test('blank workflow names can be submitted and use the server assigned name', async ({ page }) => {
  await page.route('**/api/workflows', async route => {
    if (route.request().method() !== 'POST') return route.fulfill({ json: { workflows } });
    const body = route.request().postDataJSON();
    expect(body.name).toBe('');
    return route.fulfill({ json: { Workflow: { ...workflows[0], id: 'unnamed', name: 'untitled_1' }, Version: { graph: body.graph } } });
  });
  await page.goto('/#/workflows');
  await page.getByRole('button', { name: 'Add Workflow' }).click();
  await page.getByRole('button', { name: 'Create & Open Canvas' }).click();
  await expect(page.getByLabel('Workflow name', { exact: true })).toHaveValue('untitled_1');
});


test('blank names work with older backends that require a name', async ({ page }) => {
  const names: string[] = [];
  await page.route('**/api/workflows', async route => {
    if (route.request().method() !== 'POST') return route.fulfill({ json: { workflows: [...workflows, { ...workflows[0], id: 'existing-untitled', name: 'untitled_2' }] } });
    const body = route.request().postDataJSON();
    names.push(body.name);
    if (!body.name) return route.fulfill({ status: 400, json: { error: 'Name is required' } });
    return route.fulfill({ json: { Workflow: { ...workflows[0], id: 'unnamed', name: body.name }, Version: { graph: body.graph } } });
  });
  await page.goto('/#/workflows');
  await page.getByRole('button', { name: 'Add Workflow' }).click();
  await page.getByLabel('Workflow Name (optional)', { exact: true }).fill('   ');
  await page.getByRole('button', { name: 'Create & Open Canvas' }).click();
  await expect(page.getByLabel('Workflow name', { exact: true })).toHaveValue('untitled_3');
  expect(names).toEqual(['', 'untitled_3']);
});
