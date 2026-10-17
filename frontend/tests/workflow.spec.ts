import { test, expect } from '@playwright/test';
import { authenticateLocalWorkspace } from './fixtures/local-auth';

test.use({ storageState: process.env.CLERK_TEST_STORAGE_STATE });

test('create, save, reload, execute, schedule and sign a webhook', async ({ page, request }) => {
  if (process.env.E2E_LOCAL_AUTH === '1') {
    await authenticateLocalWorkspace(page, request);
  } else if (!process.env.CLERK_TEST_STORAGE_STATE) {
    throw new Error('Set E2E_LOCAL_AUTH=1 for a local JWT backend, or provide CLERK_TEST_STORAGE_STATE for Clerk.');
  }
  const name = `Browser workflow ${Date.now()}`;
  await page.goto('/');
  await page.getByRole('button', { name: 'Add Workflow' }).click();
  await page.getByPlaceholder('e.g. GitHub Webhook Notifier').fill(name);
  await page.getByRole('button', { name: 'Create & Open Canvas' }).click();
  await expect(page.getByPlaceholder('Workflow Name')).toHaveValue(name);
  await expect(page.locator('.react-flow__node')).toHaveCount(1);
  await expect(page.locator('.react-flow__edge')).toHaveCount(0);
  await page.getByPlaceholder('Add description...').fill('Persisted from the browser');
  await page.getByRole('button', { name: 'Save Graph' }).click();
  await expect(page.getByText('Workflow saved successfully!')).toBeVisible();
  await page.reload();
  await page.getByRole('link', { name: 'Workflows', exact: true }).click();
  await page.getByRole('heading', { name, exact: true }).click();
  await expect(page.getByPlaceholder('Add description...')).toHaveValue('Persisted from the browser');
  await expect(page.locator('.react-flow__node')).toHaveCount(1);
  await page.getByRole('button', { name: 'Execute Workflow' }).click();
  await expect(page.getByText('Execution SUCCEEDED', {exact: true})).toBeVisible({timeout: 20_000});
  await page.getByTitle('Back to Dashboard').click();
  await page.getByRole('link', { name: 'Executions', exact: true }).click();
  await expect(page.getByText('Succeeded', {exact: true})).toBeVisible();
  await page.getByRole('link', { name: 'Schedules', exact: true }).click();
  await page.getByRole('button', {name: 'Activate Schedule'}).click();
  await expect(page.getByText('Active Schedules (1)')).toBeVisible();
  await page.reload();
  await page.getByRole('link', { name: 'Schedules', exact: true }).click();
  await expect(page.getByText('Active Schedules (1)')).toBeVisible();
  await page.getByTitle('Deactivate Schedule').click();
  await expect(page.getByText('Active Schedules (0)')).toBeVisible();
  await page.getByRole('link', { name: 'Webhooks', exact: true }).click();
  const response = page.waitForResponse(r => r.url().includes('/webhook/') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Send Payload' }).click();
  expect((await response).status()).toBe(200);
  await page.getByRole('link', { name: 'Workflows', exact: true }).click();
  await expect(page.getByRole('heading', { name, exact: true })).toHaveCount(1);
  await page.getByRole('heading', { name, exact: true }).click();
  await page.route('**/api/workflows/*', route => route.request().method() === 'PUT'
    ? route.fulfill({status: 503, contentType: 'application/json', body: '{"error":"Database unavailable"}'})
    : route.continue());
  await page.getByRole('button', { name: 'Save Graph' }).click();
  await expect(page.getByText('Error saving: Database unavailable')).toBeVisible();
  await expect(page.getByText('Workflow saved successfully!')).not.toBeVisible();
});
