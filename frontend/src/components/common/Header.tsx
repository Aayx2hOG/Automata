import React from 'react';
import { Bot, Layers, PlayCircle, Clock, Webhook, LogOut, ShieldCheck, Plus, Home } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface HeaderProps {
  activeTab: 'home' | 'workflows' | 'runs' | 'schedules' | 'webhooks';
  setActiveTab: (tab: 'home' | 'workflows' | 'runs' | 'schedules' | 'webhooks') => void;
  onNewWorkflow: () => void;
}

export const Header: React.FC<HeaderProps> = ({ activeTab, setActiveTab, onNewWorkflow }) => {
  const { user, logout } = useAuth();

  return (
    <header className="h-16 border-b border-white/10 bg-[#10141d] backdrop-blur-xl sticky top-0 z-30 px-6 flex items-center justify-between">
      {/* Brand Logo */}
      <div className="flex items-center space-x-8">
        <div className="flex items-center space-x-3 cursor-pointer" onClick={() => setActiveTab('home')}>
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-[#ff6d5a] to-[#ea4e43] flex items-center justify-center shadow-lg shadow-[#ff6d5a]/25">
            <Bot size={20} className="text-white" />
          </div>
          <div>
            <span className="font-extrabold text-base text-white tracking-wide">automata</span>
            <span className="ml-2 text-[10px] font-mono px-2 py-0.5 rounded-full bg-[#ff6d5a]/10 text-[#ff6d5a] border border-[#ff6d5a]/30 font-semibold">
              n8n-Engine
            </span>
          </div>
        </div>

        {/* Simplified Navigation Tabs */}
        <nav className="hidden md:flex items-center space-x-1">
          <button
            onClick={() => setActiveTab('home')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-colors ${
              activeTab === 'home'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Home size={15} />
            <span>Overview</span>
          </button>

          <button
            onClick={() => setActiveTab('workflows')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-colors ${
              activeTab === 'workflows'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Layers size={15} />
            <span>Workflows</span>
          </button>

          <button
            onClick={() => setActiveTab('runs')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-colors ${
              activeTab === 'runs'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <PlayCircle size={15} />
            <span>Executions</span>
          </button>

          <button
            onClick={() => setActiveTab('schedules')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-colors ${
              activeTab === 'schedules'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Clock size={15} />
            <span>Schedules</span>
          </button>

          <button
            onClick={() => setActiveTab('webhooks')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-colors ${
              activeTab === 'webhooks'
                ? 'bg-white/10 text-white shadow-inner'
                : 'text-gray-400 hover:text-gray-200 hover:bg-white/5'
            }`}
          >
            <Webhook size={15} />
            <span>Webhooks</span>
          </button>
        </nav>
      </div>

      {/* User Actions & Add Workflow button */}
      <div className="flex items-center space-x-3">
        <div className="hidden sm:flex items-center space-x-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-[11px] font-medium">
          <ShieldCheck size={13} />
          <span>Active</span>
        </div>

        <button
          onClick={onNewWorkflow}
          className="flex items-center space-x-1.5 px-4 py-2 rounded-xl bg-gradient-to-r from-[#ff6d5a] to-[#ea4e43] hover:from-[#ff8575] hover:to-[#ff6d5a] text-white text-xs font-bold shadow-lg shadow-[#ff6d5a]/25 transition-all hover:scale-105 active:scale-95"
        >
          <Plus size={16} />
          <span>Add Workflow</span>
        </button>

        {user && (
          <div className="flex items-center space-x-2 pl-2 border-l border-white/10">
            <div className="text-right hidden lg:block">
              <div className="text-xs font-semibold text-white">{user.email}</div>
            </div>
            <button
              onClick={logout}
              className="p-2 rounded-xl text-gray-400 hover:bg-rose-500/10 hover:text-rose-400 transition-colors"
              title="Sign Out"
            >
              <LogOut size={16} />
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
