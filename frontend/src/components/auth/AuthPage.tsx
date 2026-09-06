import React, { useState } from 'react';
import { Bot, Sparkles, Lock, Mail, ArrowRight, ShieldCheck, Zap } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

export const AuthPage: React.FC = () => {
  const { login, register, loginAsDemo, error, clearError } = useAuth();
  const [isRegister, setIsRegister] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) return;

    setIsSubmitting(true);
    try {
      if (isRegister) {
        await register(email, password);
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
    <div className="min-h-screen bg-[#0b0f19] text-gray-100 flex items-center justify-center p-4 relative overflow-hidden">
      {/* Background Ambient Glow FX */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-indigo-600/20 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-600/20 rounded-full blur-3xl pointer-events-none" />

      <div className="max-w-md w-full relative z-10 space-y-8">
        {/* Brand Header */}
        <div className="text-center space-y-3">
          <div className="inline-flex p-3 rounded-2xl bg-gradient-to-tr from-indigo-600 to-purple-600 shadow-2xl shadow-indigo-500/40 mb-2">
            <Bot size={36} className="text-white" />
          </div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Automata</h1>
          <p className="text-sm text-gray-400">
            Durable workflow automation engine & visual builder
          </p>
        </div>

        {/* Auth Card */}
        <div className="bg-slate-900/80 border border-white/10 rounded-3xl p-8 backdrop-blur-2xl shadow-2xl space-y-6">
          {/* Tab Selector */}
          <div className="flex bg-slate-950/70 p-1 rounded-xl border border-white/5">
            <button
              onClick={() => {
                setIsRegister(false);
                clearError();
              }}
              className={`flex-1 py-2 text-xs font-semibold rounded-lg transition-all ${
                !isRegister ? 'bg-indigo-600 text-white shadow' : 'text-gray-400 hover:text-white'
              }`}
            >
              Sign In
            </button>
            <button
              onClick={() => {
                setIsRegister(true);
                clearError();
              }}
              className={`flex-1 py-2 text-xs font-semibold rounded-lg transition-all ${
                isRegister ? 'bg-indigo-600 text-white shadow' : 'text-gray-400 hover:text-white'
              }`}
            >
              Create Account
            </button>
          </div>

          {error && (
            <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs text-center font-medium">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="text-xs font-semibold text-gray-300 uppercase tracking-wider">Email Address</label>
              <div className="mt-1 relative">
                <Mail className="absolute left-3.5 top-3 text-gray-500" size={18} />
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="admin@automata.io"
                  className="w-full bg-slate-950 border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500 transition-colors"
                />
              </div>
            </div>

            <div>
              <label className="text-xs font-semibold text-gray-300 uppercase tracking-wider">Password</label>
              <div className="mt-1 relative">
                <Lock className="absolute left-3.5 top-3 text-gray-500" size={18} />
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                  className="w-full bg-slate-950 border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500 transition-colors"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold text-sm shadow-lg shadow-indigo-600/30 transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 flex items-center justify-center space-x-2"
            >
              <span>{isSubmitting ? 'Authenticating...' : isRegister ? 'Register Account' : 'Sign In'}</span>
              <ArrowRight size={16} />
            </button>
          </form>

          <div className="relative flex py-1 items-center">
            <div className="flex-grow border-t border-white/10" />
            <span className="flex-shrink mx-4 text-[10px] text-gray-500 uppercase tracking-wider">or</span>
            <div className="flex-grow border-t border-white/10" />
          </div>

          {/* Quick Demo Mode button */}
          <button
            onClick={loginAsDemo}
            className="w-full py-2.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-indigo-300 font-medium text-xs transition-colors flex items-center justify-center space-x-2 group"
          >
            <Zap size={14} className="text-amber-400 group-hover:scale-125 transition-transform" />
            <span>Launch Quick Demo Mode</span>
          </button>
        </div>

        {/* Footer features */}
        <div className="flex items-center justify-center space-x-6 text-[11px] text-gray-500 font-medium">
          <div className="flex items-center space-x-1">
            <ShieldCheck size={14} className="text-emerald-400" />
            <span>Argon2id & JWT Auth</span>
          </div>
          <div className="flex items-center space-x-1">
            <Sparkles size={14} className="text-indigo-400" />
            <span>React Flow Canvas</span>
          </div>
        </div>
      </div>
    </div>
  );
};
