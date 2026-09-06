import React from 'react';
import { Bot, Layers, PlayCircle, Clock, Webhook, LogOut, ShieldCheck, Sparkles } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface HeaderProps {
  activeTab: 'workflows' | 'runs' | 'schedules' | 'webhooks';
  setActiveTab: (tab: 'workflows' | 'runs' | 'schedules' | 'webhooks') => void;
  onNewWorkflow: () => void;
}

export const Header: React.FC<HeaderProps> = ({ activeTab, setActiveTab, onNewWorkflow }) => {
  const { user, logout } = useAuth();

  return (
    <header className="h-16 border-b border-white/10 bg-slate-950/80 backdrop-blur-xl sticky top-0 z-30 px-6 flex items-center justify-between">
      {/* Brand Logo */}
      <div className="flex items-center space-x-8">
        <div className="flex items-center space-x-3 cursor-pointer" onClick={() => setActiveTab('workflows')}>
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-purple-500 flex items-center justify-center shadow-lg shadow-indigo-500/25">
            <Bot size={22} className="text-white" />
          </div>
          <div>
            <span className="font-bold text-lg text-white tracking-wide">Automata</span>
            <span className="ml-2 text-[10px] font-mono px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
              v1.0
            </span>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="hidden md:flex items-center space-x-1">
          <button
            onClick={() => setActiveTab('workflows')}
            className={`flex items-center space-x-2 px-3 py-2 rounded-xl text-sm font-medium transition-colors ${
              activeTab === 'workflows'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Layers size={16} />
            <span>Workflows</span>
          </button>

          <button
            onClick={() => setActiveTab('runs')}
            className={`flex items-center space-x-2 px-3 py-2 rounded-xl text-sm font-medium transition-colors ${
              activeTab === 'runs'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <PlayCircle size={16} />
            <span>Executions</span>
          </button>

          <button
            onClick={() => setActiveTab('schedules')}
            className={`flex items-center space-x-2 px-3 py-2 rounded-xl text-sm font-medium transition-colors ${
              activeTab === 'schedules'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Clock size={16} />
            <span>Schedules</span>
          </button>

          <button
            onClick={() => setActiveTab('webhooks')}
            className={`flex items-center space-x-2 px-3 py-2 rounded-xl text-sm font-medium transition-colors ${
              activeTab === 'webhooks'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Webhook size={16} />
            <span>Webhooks</span>
          </button>
        </nav>
      </div>

      {/* User Actions */}
      <div className="flex items-center space-x-4">
        {/* Backend health status badge */}
        <div className="hidden sm:flex items-center space-x-1.5 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
          <ShieldCheck size={14} />
          <span>Engine Online</span>
        </div>

        <button
          onClick={onNewWorkflow}
          className="flex items-center space-x-2 px-4 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white text-sm font-semibold shadow-lg shadow-indigo-600/25 transition-all hover:scale-105 active:scale-95"
        >
          <Sparkles size={16} />
          <span>New Workflow</span>
        </button>

        {user && (
          <div className="flex items-center space-x-3 pl-2 border-l border-white/10">
            <div className="text-right hidden lg:block">
              <div className="text-xs font-semibold text-white">{user.email}</div>
              <div className="text-[10px] text-gray-400">Authenticated</div>
            </div>
            <button
              onClick={logout}
              className="p-2 rounded-xl text-gray-400 hover:bg-rose-500/10 hover:text-rose-400 transition-colors"
              title="Sign Out"
            >
              <LogOut size={18} />
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
