import React, { useState } from 'react';
import { Bot, Sparkles, Lock, Mail, ArrowRight, ShieldCheck, UserRound, Check } from 'lucide-react';
import { useAuth } from '../../context/useAuth';

export const AuthPage: React.FC = () => {
  const { login, register, error, clearError } = useAuth();
  const [isRegister, setIsRegister] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password || (isRegister && (!username || !confirmPassword))) return;
    if (isRegister && password !== confirmPassword) {
      setFormError('Passwords do not match.');
      return;
    }

    setFormError(null);
    setIsSubmitting(true);
    try {
      if (isRegister) {
        await register(username, email, password);
      } else {
        await login(email, password);
      }
    } catch {
      // Handled in context
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#0b0e14] text-gray-100 flex items-center justify-center p-4 relative overflow-hidden font-sans">
      {/* Background Ambient Coral Glow */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-[#ff6d5a]/10 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-[#ea4e43]/10 rounded-full blur-3xl pointer-events-none" />

      <div className="max-w-md w-full relative z-10 space-y-8">
        {/* Brand Header */}
        <div className="text-center space-y-3">
          <div className="inline-flex p-3 rounded-2xl bg-gradient-to-tr from-[#ff6d5a] to-[#ea4e43] shadow-2xl shadow-[#ff6d5a]/40 mb-2">
            <Bot size={36} className="text-white" />
          </div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">automata</h1>
          <p className="text-xs text-gray-400">
            Durable n8n-style workflow automation & execution engine
          </p>
        </div>

        {/* Auth Card */}
        <div className="bg-[#161a23] border border-white/10 rounded-3xl p-8 backdrop-blur-2xl shadow-2xl space-y-6">
          {/* Tab Selector */}
          <div className="flex bg-[#0d1017] p-1 rounded-xl border border-white/5">
            <button
              onClick={() => {
                setIsRegister(false);
                setFormError(null);
                clearError();
              }}
              className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all ${
                !isRegister ? 'bg-[#ff6d5a] text-white shadow' : 'text-gray-400 hover:text-white'
              }`}
            >
              Sign In
            </button>
            <button
              onClick={() => {
                setIsRegister(true);
                setFormError(null);
                clearError();
              }}
              className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all ${
                isRegister ? 'bg-[#ff6d5a] text-white shadow' : 'text-gray-400 hover:text-white'
              }`}
            >
              Create Account
            </button>
          </div>

          {error && (
            <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs text-center font-semibold">
              {error}
            </div>
          )}
          {formError && (
            <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs text-center font-semibold">
              {formError}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {isRegister && (
              <div>
                <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Username</label>
                <div className="mt-1 relative">
                  <UserRound className="absolute left-3.5 top-3 text-gray-500" size={18} />
                  <input
                    type="text"
                    required
                    minLength={3}
                    maxLength={30}
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="your-workspace-name"
                    autoComplete="username"
                    className="w-full bg-[#0d1017] border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
                  />
                </div>
                <p className="mt-1 text-[11px] text-gray-500">3–30 characters, used to identify you in Automata.</p>
              </div>
            )}
            <div>
              <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Email Address</label>
              <div className="mt-1 relative">
                <Mail className="absolute left-3.5 top-3 text-gray-500" size={18} />
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="admin@automata.io"
                  autoComplete="email"
                  className="w-full bg-[#0d1017] border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
                />
              </div>
            </div>

            <div>
              <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Password</label>
              <div className="mt-1 relative">
                <Lock className="absolute left-3.5 top-3 text-gray-500" size={18} />
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                  autoComplete={isRegister ? 'new-password' : 'current-password'}
                  className="w-full bg-[#0d1017] border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
                />
              </div>
            </div>

            {isRegister && (
              <div>
                <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Confirm Password</label>
                <div className="mt-1 relative">
                  <Check className="absolute left-3.5 top-3 text-gray-500" size={18} />
                  <input
                    type="password"
                    required
                    value={confirmPassword}
                    onChange={(e) => {
                      setConfirmPassword(e.target.value);
                      setFormError(null);
                    }}
                    placeholder="Re-enter your password"
                    autoComplete="new-password"
                    className="w-full bg-[#0d1017] border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
                  />
                </div>
                {confirmPassword && password !== confirmPassword && (
                  <p className="mt-1 text-[11px] text-rose-300">Passwords do not match.</p>
                )}
              </div>
            )}

            {isRegister && (
              <p className="text-[11px] leading-relaxed text-gray-500">
                Create an account to build, schedule, and monitor your automations from one workspace.
              </p>
            )}

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 rounded-xl bg-gradient-to-r from-[#ff6d5a] to-[#ea4e43] hover:from-[#ff8575] hover:to-[#ff6d5a] text-white font-bold text-sm shadow-lg shadow-[#ff6d5a]/30 transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 flex items-center justify-center space-x-2"
            >
              <span>{isSubmitting ? (isRegister ? 'Creating account...' : 'Signing in...') : isRegister ? 'Create Account' : 'Sign In'}</span>
              <ArrowRight size={16} />
            </button>
          </form>

        </div>

        {/* Footer features */}
        <div className="flex items-center justify-center space-x-6 text-[11px] text-gray-500 font-medium">
          <div className="flex items-center space-x-1">
            <ShieldCheck size={14} className="text-emerald-400" />
            <span>Argon2id & JWT Auth</span>
          </div>
          <div className="flex items-center space-x-1">
            <Sparkles size={14} className="text-[#ff6d5a]" />
            <span>n8n Visual Builder</span>
          </div>
        </div>
      </div>
    </div>
  );
};
