import { test, expect } from '@playwright/test';
import { passwordStrength } from '../src/lib/password-strength';

test('classifier recognizes common passwords, length, and stronger phrases', () => {
  expect(passwordStrength('12345678').label).toBe('Weak');
  expect(passwordStrength('aZ7!q2').label).toBe('Weak');
  expect(passwordStrength('abc123xyz!').label).toBe('Medium');
  expect(passwordStrength('brisk-otter-lantern-orbit-739!').label).toBe('Strong');
  expect(passwordStrength('passwordpasswordpassword').label).toBe('Weak');
});

test('password strength updates, clears, and survives visibility toggles', async ({ page }) => {
  // Exercise the wrapper without a live Clerk account or real passwords.
  await page.route('**/__password-strength', route => route.fulfill({
    contentType: 'text/html',
    body: `<div id="root"></div><script type="module">
      import RefreshRuntime from '/@react-refresh';
      RefreshRuntime.injectIntoGlobalHook(window);
      window.$RefreshReg$ = () => {};
      window.$RefreshSig$ = () => type => type;
      window.__vite_plugin_react_preamble_installed__ = true;
      const React = await import('/node_modules/.vite/deps/react.js');
      const { createRoot } = await import('/node_modules/.vite/deps/react-dom_client.js');
      const { PasswordStrength } = await import('/src/components/auth/PasswordStrength.tsx');
      createRoot(document.getElementById('root')).render(React.createElement(PasswordStrength, null,
        React.createElement('input', { type: 'password', name: 'password', 'aria-label': 'Password' })));
    </script>`,
  }));
  await page.goto('/__password-strength');
  const password = page.getByLabel('Password');
  const status = page.getByRole('status');
  await password.fill('12345678');
  await expect(status).toHaveText('Password strength: Weak');
  await password.fill('abc123xyz!');
  await expect(status).toHaveText('Password strength: Medium');
  await password.evaluate(input => { (input as HTMLInputElement).type = 'text'; });
  await password.fill('brisk-otter-lantern-orbit-739!');
  await expect(status).toHaveText('Password strength: Strong');
  await password.fill('');
  await expect(status).toHaveCount(0);
  await password.fill('short');
  await expect(status).toHaveText('Password strength: Weak');
  await password.evaluate(input => input.remove());
  await expect(status).toHaveCount(0);
});
