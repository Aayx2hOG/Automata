import { useState } from 'react';
import { SignIn, SignUp, UserButton, useAuth } from '@clerk/react';
import { ArrowRight, Check, Clock3, GitBranch, Globe2, LockKeyhole, RotateCcw, Webhook, Workflow } from 'lucide-react';
import './auth.css';
import { BrandMark } from '../common/BrandMark';
import { PasswordStrength } from './PasswordStrength';

interface AuthPageProps {
  pending?: boolean;
  error?: string | null;
  onRetry?: () => void;
}

export const AuthPage = ({ pending = false, error, onRetry }: AuthPageProps) => {
  const [mode, setMode] = useState<'signin' | 'signup'>('signin');
  const { isLoaded, isSignedIn } = useAuth();
  const appearance = {
    elements: {
      rootBox: { width: '100%' },
      cardBox: { width: '100%', boxShadow: 'none', border: 'none', borderRadius: '16px' },
      card: { background: 'transparent', boxShadow: 'none', padding: '28px' },
      headerTitle: { fontSize: '25px', letterSpacing: '-0.7px', fontWeight: '600' },
      headerSubtitle: { fontSize: '13px', lineHeight: '1.6' },
      socialButtonsBlockButton: { minHeight: '46px', border: '1px solid #363a40', background: '#242427' },
      formButtonPrimary: { minHeight: '44px', boxShadow: 'none' },
      formFieldInput: { minHeight: '44px', border: '1px solid #363a40', boxShadow: 'none' },
      footer: { background: '#19191c', backgroundImage: 'none' },
    },
  };

  return (
    <div className="auth-page">
      <header className="auth-nav">
        <a href="#/welcome" className="auth-brand" aria-label="Automata home"><span className="auth-brand-icon"><BrandMark size={34} /></span>Automata</a>
        <nav aria-label="Account" className="auth-nav-actions">
          {isSignedIn ? <><a className="auth-nav-signup" href="#/home">Open workspace <ArrowRight size={14} /></a><UserButton /></> : <>
            <button className={mode === 'signin' ? 'auth-nav-link active' : 'auth-nav-link'} onClick={() => setMode('signin')}>Sign in</button>
            <button className="auth-nav-signup" onClick={() => setMode('signup')}>Sign up <ArrowRight size={14} /></button>
          </>}
        </nav>
      </header>

      <main className="auth-main">
        <section className="auth-story" aria-labelledby="auth-title">
          <div className="auth-eyebrow"><span /> YOUR WORK, CONNECTED</div>
          <h1 id="auth-title">Make the work<br />flow<span className="auth-title-accent">.</span></h1>
          <p className="auth-description">Turn everyday tasks into dependable workflows.<br className="auth-desktop-break" /> Build visually. Run on your schedule. Stay in control.</p>

          <div className="auth-workflow-preview" aria-label="Example workflow: a webhook triggers an API request, then a condition checks the response">
            <div className="auth-preview-bar"><span><Workflow size={14} /> Lead enrichment</span><span className="auth-preview-label">EXAMPLE WORKFLOW</span></div>
            <div className="auth-canvas">
              <div className="auth-node"><div className="auth-node-icon coral"><Webhook size={21} /></div><strong>New lead</strong><span>Webhook trigger</span><i className="auth-port out" /></div>
              <div className="auth-connection"><ArrowRight size={12} /></div>
              <div className="auth-node"><div className="auth-node-icon blue"><Globe2 size={21} /></div><strong>Enrich data</strong><span>HTTP request</span><i className="auth-port in" /><i className="auth-port out" /></div>
              <div className="auth-connection"><ArrowRight size={12} /></div>
              <div className="auth-node"><div className="auth-node-icon green"><GitBranch size={21} /></div><strong>Check result</strong><span>Condition</span><i className="auth-port in" /></div>
            </div>
            <div className="auth-preview-footer"><span><Check size={13} /> A clear path from trigger to result</span><span>3 nodes</span></div>
          </div>

          <div className="auth-features">
            <div><Workflow size={17} /><span>Visual builder</span></div>
            <div><Clock3 size={17} /><span>Scheduled runs</span></div>
            <div><RotateCcw size={17} /><span>Durable execution</span></div>
          </div>
        </section>

        <section className="auth-entry" aria-label="Your Automata account" aria-busy={pending}>
          <div className="auth-panel">
            <div className="auth-panel-overline"><span className="auth-panel-mark"><BrandMark size={22} /></span> YOUR WORKSPACE AWAITS</div>
            {error ? <div className="auth-account-error">
              <h2>Let’s reconnect.</h2>
              <p role="alert">{error}</p>
              <button className="auth-retry" onClick={onRetry}>Try again <ArrowRight size={16} /></button>
              <p className="auth-error-hint">You are signed in. You can manage your account from the profile button above.</p>
            </div> : pending || !isLoaded ? <div className="auth-placeholder" role="status" aria-label="Preparing your workspace">
              <div className="auth-placeholder-title" /><div className="auth-placeholder-subtitle" /><div className="auth-placeholder-button" /><div className="auth-placeholder-divider" /><div className="auth-placeholder-button" /><div className="auth-placeholder-button" />
            </div> : isSignedIn ? <div className="auth-welcome">
              <h2>Welcome back.</h2>
              <p>Your workflows are ready when you are.</p>
              <a className="auth-workspace-button" href="#/home">Open workspace <ArrowRight size={17} aria-hidden="true" /></a>
            </div> : <PasswordStrength key={mode}>{mode === 'signin' ? <SignIn routing="hash" appearance={appearance} forceRedirectUrl="/" /> : <SignUp routing="hash" appearance={appearance} forceRedirectUrl="/" />}</PasswordStrength>}
          </div>
          {!isSignedIn && <p className="auth-security"><LockKeyhole size={12} /> Secure access with Google or email</p>}
        </section>
      </main>
      <footer className="auth-footer"><span>Less repetition. More creation.</span><span>Built for the way you work.</span></footer>
    </div>
  );
};
