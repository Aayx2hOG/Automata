import { useEffect, useRef, useState, type ReactNode } from 'react';

// Clerk owns password validation. Keep only the score here, never the password.
export function PasswordStrength({ children }: { children: ReactNode }) {
  const root = useRef<HTMLDivElement>(null);
  const field = useRef<HTMLInputElement | null>(null);
  const revision = useRef(0);
  const [strength, setStrength] = useState<{ label: string; level: number } | null>(null);

  useEffect(() => {
    const observer = new MutationObserver(() => {
      if (field.current && !root.current?.contains(field.current)) {
        field.current = null;
        revision.current++;
        setStrength(null);
      }
    });
    if (root.current) observer.observe(root.current, { childList: true, subtree: true });
    return () => observer.disconnect();
  }, []);

  return <div ref={root} onInputCapture={event => {
    const input = event.target;
    if (!(input instanceof HTMLInputElement) ||
        !(input.type === 'password' || /password/i.test(input.name))) return;
    field.current = input;
    const currentRevision = ++revision.current;
    if (!input.value) { setStrength(null); return; }
    void import('../../lib/password-strength').then(({ passwordStrength }) => {
      if (currentRevision === revision.current && root.current?.contains(input)) {
        setStrength(passwordStrength(input.value));
      }
    }).catch(() => {
      // A failed optional chunk must not prevent Clerk from accepting the form.
      if (currentRevision === revision.current) setStrength(null);
    });
  }}>
    {children}
    {strength && <div className="auth-password-strength" data-level={strength.level} role="status" aria-live="polite" aria-atomic="true">
      <div className="auth-password-strength-bars" aria-hidden="true">
        {[1, 2, 3].map(level => <span key={level} data-filled={level <= strength.level} />)}
      </div>
      <span>Password strength: <strong>{strength.label}</strong></span>
    </div>}
  </div>;
}
