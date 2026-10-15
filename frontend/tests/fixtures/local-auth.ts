import { expect, Page, APIRequestContext } from '@playwright/test';

export async function authenticateLocalWorkspace(page: Page, request: APIRequestContext) {
  const id = crypto.randomUUID();
  const response = await request.post('/api/auth/register', {
    data: { username: `demo_${id.slice(0, 8)}`, email: `demo_${id}@example.com`, password: `Test-${id}!` },
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  const tokens = await response.json();
  await page.addInitScript(tokens => {
    localStorage.setItem('automata_access_token', tokens.access_token);
    localStorage.setItem('automata_refresh_token', tokens.refresh_token);
  }, tokens);
  await page.route('**/src/main.tsx', route => route.fulfill({
    contentType: 'application/javascript', body: 'import "/tests/fixtures/workspace.tsx";',
  }));
  // Only replace the provider-owned profile widget, not application authentication or API calls.
  await page.route('**/src/components/common/Header.tsx', async route => {
    const response = await route.fetch();
    const source = await response.text();
    // Vite has transformed the import by the time the browser requests this module.
    const transformed = source.replace(/import \{ UserButton \} from [^;]+;/, 'const UserButton = () => null;');
    expect(transformed).not.toBe(source);
    await route.fulfill({ response, body: transformed });
  });
}
