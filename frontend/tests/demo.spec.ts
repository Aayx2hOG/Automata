import { expect, test } from '@playwright/test';
import { authenticateLocalWorkspace } from './fixtures/local-auth';

// Opt-in capture against a dedicated, migrated local JWT backend.
test.use({ video: { mode: 'on', size: { width: 1440, height: 1000 } } });

test('record a real workflow demo', async ({ page, request }) => {
  test.skip(process.env.RECORD_DEMO !== '1', 'Run npm run demo against a local test backend.');
  await authenticateLocalWorkspace(page, request);
  await page.goto('/');
  await expect(page.getByText('Your first workflow starts here')).toBeVisible();
  await page.getByRole('button', { name: 'Add Workflow' }).click();
  await page.getByLabel('Workflow Name').fill('Trigger check → audit log');
  await page.getByLabel('Description', { exact: true }).fill('Check a trigger and record the selected branch.');
  await page.getByRole('button', { name: 'Create & Open Canvas' }).click();
  await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();

  // Seed a useful graph through the same authenticated API used by the editor.
  await page.evaluate(async () => {
    const { api } = await import('/src/lib/api.ts');
    const workflows = await api.workflows.list();
    await api.workflows.update(workflows[0].id, workflows[0].name, workflows[0].description || '', {
      nodes: [
        { id: 'start', type: 'manual_trigger', label: 'Start audit', config: {}, position: { x: 250, y: 40 } },
        { id: 'check', type: 'condition', label: 'Trigger received?', config: { source_node: 'start', source_field: 'triggered', operator: 'equals', value: 'true' }, position: { x: 250, y: 210 } },
        { id: 'accepted', type: 'logger', label: 'Record success', config: { message: 'Trigger verified' }, position: { x: 50, y: 400 } },
        { id: 'rejected', type: 'logger', label: 'Record rejection', config: { message: 'Trigger missing' }, position: { x: 450, y: 400 } },
      ],
      edges: [{ from_node_id: 'start', to_node_id: 'check' }, { from_node_id: 'check', to_node_id: 'accepted', condition: 'true' }, { from_node_id: 'check', to_node_id: 'rejected', condition: 'false' }],
    });
  });
  await page.reload();
  await page.getByRole('button', { name: /Trigger check → audit log/ }).click();
  await expect(page.locator('.react-flow__node')).toHaveCount(4);
  await page.getByLabel('Workflow description').fill('Route a trigger and inspect the checkpoint for each completed step.');
  await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Save Graph' }).click();
  await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Execute Workflow' }).click();
  await expect(page.getByText('Execution SUCCEEDED', { exact: true })).toBeVisible({ timeout: 20000 });
  await page.getByLabel('Node executions').locator('summary').first().click();
  await expect(page.getByText('Workflow execution triggered!', { exact: true })).not.toBeVisible();
  await page.screenshot({ path: '../docs/demo/editor.png' });
  await page.getByTitle('Back to Dashboard').click();
  await expect(page.getByRole('heading', { name: 'Recent workflows' })).toBeVisible();
  await expect(page.getByText('succeeded', { exact: true })).toBeVisible();
  await page.screenshot({ path: '../docs/demo/workspace.png', fullPage: true });
  await page.getByRole('button', { name: 'View execution history' }).click();
  await expect(page.getByText('Succeeded', { exact: true })).toBeVisible();
});
