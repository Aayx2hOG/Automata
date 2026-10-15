import { test, expect } from '@playwright/test';
import { WorkflowGraph } from '../src/types';

test.beforeEach(async ({ page }) => {
  await page.route('**/src/main.tsx', route => route.fulfill({ contentType: 'application/javascript', body: 'import "/tests/fixtures/editor.tsx";' }));
  await page.route('**/api/credentials', route => route.fulfill({ json: [{ id: 'credential-one', name: 'Example API', host: 'api.example.com', header: 'Authorization', updated_at: new Date().toISOString() }] }));
  await page.route('**/test/save', route => route.fulfill({ json: {} }));
  await page.goto('/');
});

test('editor saves credential references, mappings and retry policies without secrets', async ({ page }) => {
  await page.getByText('Send customer', { exact: true }).click();
  await page.getByLabel('Stored credential').selectOption('credential-one');
  await page.getByLabel('HTTP error handling').selectOption('fail');
  await page.getByLabel('Maximum attempts').fill('3');
  await page.getByLabel('Backoff (seconds)').fill('2');
  await page.getByLabel('Destination field').selectOption('body');
  await page.getByRole('combobox', { name: 'Output field', exact: true }).selectOption('body');
  await page.getByRole('button', { name: 'Apply mapping' }).click();
  const saved = page.waitForRequest(request => request.url().endsWith('/test/save'));
  await page.getByRole('button', { name: 'Save Changes', exact: true }).click();
  const graph = (await saved).postDataJSON() as WorkflowGraph;
  const target = graph.nodes.find(node => node.id === 'target')!;
  expect(target.config.credential_id).toBe('credential-one');
  expect(target.config.error_policy).toBe('fail');
  expect(target.retry).toEqual({ max_attempts: 3, backoff_seconds: 2 });
  expect(target.inputs).toEqual({ body: { node_id: 'source', path: ['body'] } });
  expect(target.config).not.toHaveProperty('value');
});

test('execution inspector polls checkpoints and displays failures', async ({ page }) => {
  await page.route('**/api/runs/test-run', route => route.fulfill({ json: {
    id: 'test-run', status: 'failed', error_message: 'HTTP request returned status 401',
    nodes: [{ node_id: 'source', status: 'failed', attempts: 1, error_message: 'HTTP request returned status 401', started_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:01Z' }],
  } }));
  await page.getByRole('button', { name: 'Execute Workflow' }).click();
  await expect(page.getByText('Execution FAILED', { exact: true })).toBeVisible();
  const details = page.getByLabel('Node executions');
  await expect(details).toContainText('1 attempt');
  await details.locator('summary').click();
  await expect(details).toContainText('HTTP request returned status 401');
});


test('unsaved edits warn before leaving and successful saves clear the warning', async ({ page }) => {
  await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();
  await page.getByLabel('Workflow description').fill('Unsaved draft');
  await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
  page.once('dialog', dialog => dialog.dismiss());
  await page.getByTitle('Back to Dashboard').click();
  await expect(page.locator('body')).not.toHaveAttribute('data-left-editor');
  await page.getByRole('button', { name: 'Save Graph' }).click();
  await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();
  await page.getByTitle('Back to Dashboard').click();
  await expect(page.locator('body')).toHaveAttribute('data-left-editor', 'true');
});

test('a failed save retains edits and blocks execution', async ({ page }) => {
  await page.route('**/test/save', route => route.fulfill({ status: 503, json: {} }));
  await page.getByLabel('Workflow name').fill('Draft name');
  await page.getByRole('button', { name: 'Execute Workflow' }).click();
  await expect(page.getByText('Save failed — try again', { exact: true })).toBeVisible();
  await expect(page.getByText('Execution RUNNING', { exact: true })).not.toBeVisible();
  page.once('dialog', dialog => dialog.accept());
  await page.getByTitle('Back to Dashboard').click();
  await expect(page.locator('body')).toHaveAttribute('data-left-editor', 'true');
});

test('edits made during a save remain unsaved after it completes', async ({ page }) => {
  let finishSave!: () => void;
  const gate = new Promise<void>(resolve => { finishSave = resolve; });
  await page.route('**/test/save', async route => { await gate; await route.fulfill({ json: {} }); });
  await page.getByLabel('Workflow description').fill('First draft');
  await page.getByRole('button', { name: 'Save Graph' }).click();
  await expect(page.getByText('Saving…', { exact: true })).toBeVisible();
  await page.getByLabel('Workflow description').fill('Second draft');
  finishSave();
  await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
});
