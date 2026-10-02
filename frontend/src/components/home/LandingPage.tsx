import React, { useState } from 'react';
import {
  ArrowRight,
  Webhook,
  Clock,
  Globe,
  GitFork,
  Terminal,
  ShieldCheck,
  Zap,
  CheckCircle2,
  Layers,
  Code2,
  Activity,
  Cpu,
  Plus,
} from 'lucide-react';
import { Workflow } from '../../types';

interface LandingPageProps {
  workflows: Workflow[];
  onLaunchStudio: () => void;
  onOpenWorkflow: (wf: Workflow) => void;
  onNavigateTab: (tab: 'workflows' | 'runs' | 'schedules' | 'webhooks') => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({
  workflows,
  onLaunchStudio,
  onNavigateTab,
}) => {
  const [activeInteractiveNode, setActiveInteractiveNode] = useState<'webhook' | 'http' | 'condition' | 'logger'>('webhook');

  return (
    <div className="space-y-12 py-2">
      {/* Hero Section */}
      <section className="relative rounded-3xl bg-[#161a23] border border-white/10 p-8 md:p-12 shadow-2xl overflow-hidden">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 items-center relative z-10">
          {/* Left Text */}
          <div className="lg:col-span-7 space-y-6">
            <div className="inline-flex items-center space-x-2 px-3 py-1.5 rounded-full bg-[#ff6d5a]/10 border border-[#ff6d5a]/30 text-[#ff6d5a] text-xs font-bold">
              <Zap size={14} className="text-[#ff6d5a]" />
              <span>n8n-Inspired Workflow Automation Engine</span>
            </div>

            <h1 className="text-4xl md:text-5xl font-extrabold text-white tracking-tight leading-tight">
              Build & Automate Resilient DAG Node Graphs.
            </h1>

            <p className="text-base text-gray-300 leading-relaxed max-w-xl">
              Automata gives you n8n-style visual workflow building combined with Temporal-grade durable graph execution. Connect Webhooks, HTTP Requests, Conditions, and Cron triggers into clean automation pipelines.
            </p>

            {/* CTAs */}
            <div className="flex flex-wrap items-center gap-4 pt-2">
              <button
                onClick={onLaunchStudio}
                className="flex items-center space-x-2 px-6 py-3.5 rounded-xl bg-gradient-to-r from-[#ff6d5a] to-[#ea4e43] hover:from-[#ff8575] hover:to-[#ff6d5a] text-white font-bold text-sm shadow-xl shadow-[#ff6d5a]/30 transition-all hover:scale-105 active:scale-95 group"
              >
                <Plus size={18} />
                <span>Create Workflow</span>
                <ArrowRight size={16} className="group-hover:translate-x-1 transition-transform" />
              </button>

              <button
                onClick={() => onNavigateTab('webhooks')}
                className="flex items-center space-x-2 px-5 py-3.5 rounded-xl bg-[#0d1017] hover:bg-[#1a202c] text-gray-200 font-semibold text-sm border border-white/10 transition-colors"
              >
                <Webhook size={16} className="text-[#ff6d5a]" />
                <span>Webhook Endpoints</span>
              </button>
            </div>

            {/* Metrics */}
            <div className="grid grid-cols-3 gap-4 pt-6 border-t border-white/10 max-w-lg">
              <div>
                <div className="text-2xl font-extrabold text-white">{workflows.length}</div>
                <div className="text-xs text-gray-400 font-semibold">Active Workflows</div>
              </div>
              <div>
                <div className="text-2xl font-extrabold text-emerald-400">100%</div>
                <div className="text-xs text-gray-400 font-semibold">Cycle Validated</div>
              </div>
              <div>
                <div className="text-2xl font-extrabold text-[#ff6d5a]">0ms</div>
                <div className="text-xs text-gray-400 font-semibold">Trigger Latency</div>
              </div>
            </div>
          </div>

          {/* Right n8n Flow Demo Card */}
          <div className="lg:col-span-5 bg-[#0d1017] border border-white/10 rounded-2xl p-5 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-white/10 pb-3">
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 rounded-full bg-[#ff6d5a] animate-pulse" />
                <span className="text-xs font-mono font-bold text-gray-200">N8N NODE FLOW PREVIEW</span>
              </div>
              <span className="text-[10px] font-mono text-[#ff6d5a] px-2 py-0.5 rounded bg-[#ff6d5a]/10 border border-[#ff6d5a]/30 font-bold">
                Interactive
              </span>
            </div>

            {/* Interactive Node Flow Grid */}
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={() => setActiveInteractiveNode('webhook')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activeInteractiveNode === 'webhook'
                    ? 'bg-[#ff6d5a]/20 border-[#ff6d5a] ring-1 ring-[#ff6d5a]'
                    : 'bg-[#161a23] border-white/10 hover:border-white/20'
                }`}
              >
                <div className="flex items-center space-x-2 text-[#ff6d5a] font-bold text-xs mb-1">
                  <Webhook size={14} />
                  <span>1. Webhook</span>
                </div>
                <div className="text-[11px] text-gray-300 font-mono">POST /webhook/id</div>
              </button>

              <button
                onClick={() => setActiveInteractiveNode('http')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activeInteractiveNode === 'http'
                    ? 'bg-cyan-950/60 border-cyan-500 ring-1 ring-cyan-500'
                    : 'bg-[#161a23] border-white/10 hover:border-white/20'
                }`}
              >
                <div className="flex items-center space-x-2 text-cyan-400 font-bold text-xs mb-1">
                  <Globe size={14} />
                  <span>2. HTTP API</span>
                </div>
                <div className="text-[11px] text-gray-300 font-mono">GET /api/status</div>
              </button>

              <button
                onClick={() => setActiveInteractiveNode('condition')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activeInteractiveNode === 'condition'
                    ? 'bg-amber-950/60 border-amber-500 ring-1 ring-amber-500'
                    : 'bg-[#161a23] border-white/10 hover:border-white/20'
                }`}
              >
                <div className="flex items-center space-x-2 text-amber-400 font-bold text-xs mb-1">
                  <GitFork size={14} />
                  <span>3. Condition</span>
                </div>
                <div className="text-[11px] text-gray-300 font-mono">status == 200</div>
              </button>

              <button
                onClick={() => setActiveInteractiveNode('logger')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activeInteractiveNode === 'logger'
                    ? 'bg-blue-950/60 border-blue-500 ring-1 ring-blue-500'
                    : 'bg-[#161a23] border-white/10 hover:border-white/20'
                }`}
              >
                <div className="flex items-center space-x-2 text-blue-400 font-bold text-xs mb-1">
                  <Terminal size={14} />
                  <span>4. Logger</span>
                </div>
                <div className="text-[11px] text-gray-300 font-mono">Log execution</div>
              </button>
            </div>

            {/* Step Output Box */}
            <div className="p-3.5 rounded-xl bg-[#161a23] border border-white/10 font-mono text-xs text-gray-300 space-y-2">
              <div className="text-[10px] text-gray-400 uppercase tracking-wider font-sans font-bold flex items-center justify-between">
                <span>Node Step Output</span>
                <span className="text-emerald-400 flex items-center gap-1 font-bold">
                  <CheckCircle2 size={12} /> SUCCESS
                </span>
              </div>
              {activeInteractiveNode === 'webhook' && (
                <pre className="text-[#ff6d5a] overflow-x-auto">
{`{
  "event": "user.signup",
  "payload": { "id": 42, "email": "dev@automata.io" }
}`}
                </pre>
              )}
              {activeInteractiveNode === 'http' && (
                <pre className="text-cyan-300 overflow-x-auto">
{`{
  "status": 200,
  "data": { "health": "healthy", "latency_ms": 12 }
}`}
                </pre>
              )}
              {activeInteractiveNode === 'condition' && (
                <pre className="text-amber-300 overflow-x-auto">
{`{
  "evaluated": true,
  "branch": "true_path"
}`}
                </pre>
              )}
              {activeInteractiveNode === 'logger' && (
                <pre className="text-blue-300 overflow-x-auto">
{`[INFO] 2026-09-06T16:33:00Z - Workflow step executed successfully.`}
                </pre>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* Quick Action Shortcuts */}
      <section className="space-y-4">
        <h3 className="text-xs font-bold uppercase tracking-wider text-gray-400">Quick Actions</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <button
            onClick={onLaunchStudio}
            className="p-5 rounded-2xl bg-[#161a23] border border-white/10 hover:border-[#ff6d5a] text-left transition-all group shadow-xl"
          >
            <div className="p-2.5 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a] w-fit mb-3 group-hover:scale-110 transition-transform">
              <Layers size={20} />
            </div>
            <h4 className="text-sm font-bold text-white group-hover:text-[#ff6d5a]">Create Workflow</h4>
            <p className="text-xs text-gray-400 mt-1">Design visual graph DAG</p>
          </button>

          <button
            onClick={() => onNavigateTab('schedules')}
            className="p-5 rounded-2xl bg-[#161a23] border border-white/10 hover:border-[#ff6d5a] text-left transition-all group shadow-xl"
          >
            <div className="p-2.5 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a] w-fit mb-3 group-hover:scale-110 transition-transform">
              <Clock size={20} />
            </div>
            <h4 className="text-sm font-bold text-white group-hover:text-[#ff6d5a]">Cron Schedules</h4>
            <p className="text-xs text-gray-400 mt-1">Manage interval triggers</p>
          </button>

          <button
            onClick={() => onNavigateTab('webhooks')}
            className="p-5 rounded-2xl bg-[#161a23] border border-white/10 hover:border-[#ff6d5a] text-left transition-all group shadow-xl"
          >
            <div className="p-2.5 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a] w-fit mb-3 group-hover:scale-110 transition-transform">
              <Webhook size={20} />
            </div>
            <h4 className="text-sm font-bold text-white group-hover:text-[#ff6d5a]">Webhook Tester</h4>
            <p className="text-xs text-gray-400 mt-1">cURL POST runner</p>
          </button>

          <button
            onClick={() => onNavigateTab('runs')}
            className="p-5 rounded-2xl bg-[#161a23] border border-white/10 hover:border-emerald-500 text-left transition-all group shadow-xl"
          >
            <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-400 w-fit mb-3 group-hover:scale-110 transition-transform">
              <Activity size={20} />
            </div>
            <h4 className="text-sm font-bold text-white group-hover:text-emerald-300">Execution History</h4>
            <p className="text-xs text-gray-400 mt-1">View step outputs</p>
          </button>
        </div>
      </section>

      {/* Engine Architecture */}
      <section className="space-y-6">
        <div>
          <h2 className="text-lg font-bold text-white">Engine Architecture</h2>
          <p className="text-xs text-gray-400 mt-0.5">Automata Go backend & n8n visual flow builder.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="p-6 rounded-2xl bg-[#161a23] border border-white/10 space-y-3 shadow-xl">
            <div className="p-2 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a] w-fit">
              <ShieldCheck size={20} />
            </div>
            <h3 className="text-sm font-bold text-white">DAG Graph Validation</h3>
            <p className="text-xs text-gray-300 leading-relaxed">
              Every workflow version is validated for cycles and dangling node references prior to execution.
            </p>
          </div>

          <div className="p-6 rounded-2xl bg-[#161a23] border border-white/10 space-y-3 shadow-xl">
            <div className="p-2 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a] w-fit">
              <Cpu size={20} />
            </div>
            <h3 className="text-sm font-bold text-white">Go Worker Pool</h3>
            <p className="text-xs text-gray-300 leading-relaxed">
              Concurrent worker pool execution with retries, node error handling, and state logging.
            </p>
          </div>

          <div className="p-6 rounded-2xl bg-[#161a23] border border-white/10 space-y-3 shadow-xl">
            <div className="p-2 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a] w-fit">
              <Code2 size={20} />
            </div>
            <h3 className="text-sm font-bold text-white">Extensible Node Interface</h3>
            <p className="text-xs text-gray-300 leading-relaxed">
              Add new node integrations seamlessly by implementing the Go <code className="text-[#ff6d5a] font-mono">Node</code> interface.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
};
