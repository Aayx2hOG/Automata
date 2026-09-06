import React from 'react';
import { Layers, Clock, Activity, CheckCircle2 } from 'lucide-react';
import { Workflow } from '../../types';

interface MetricsOverviewProps {
  workflows: Workflow[];
  runCount?: number;
  scheduleCount?: number;
}

export const MetricsOverview: React.FC<MetricsOverviewProps> = ({ workflows, runCount = 12, scheduleCount = 3 }) => {
  const activeWorkflows = workflows.filter((w) => w.is_active).length;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      {/* Total Workflows */}
      <div className="p-5 rounded-2xl bg-gradient-to-br from-slate-900/90 to-slate-900/50 border border-white/10 backdrop-blur-md shadow-xl relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
          <Layers size={64} className="text-indigo-400" />
        </div>
        <div className="flex items-center space-x-3 text-indigo-400 mb-2">
          <div className="p-2 rounded-xl bg-indigo-500/10 border border-indigo-500/20">
            <Layers size={20} />
          </div>
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-400">Total Workflows</span>
        </div>
        <div className="text-3xl font-extrabold text-white">{workflows.length}</div>
        <div className="text-xs text-gray-400 mt-1">{activeWorkflows} active DAG pipelines</div>
      </div>

      {/* Active Schedules */}
      <div className="p-5 rounded-2xl bg-gradient-to-br from-slate-900/90 to-slate-900/50 border border-white/10 backdrop-blur-md shadow-xl relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
          <Clock size={64} className="text-amber-400" />
        </div>
        <div className="flex items-center space-x-3 text-amber-400 mb-2">
          <div className="p-2 rounded-xl bg-amber-500/10 border border-amber-500/20">
            <Clock size={20} />
          </div>
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-400">Cron Schedules</span>
        </div>
        <div className="text-3xl font-extrabold text-white">{scheduleCount}</div>
        <div className="text-xs text-gray-400 mt-1">Automated interval triggers</div>
      </div>

      {/* Executions */}
      <div className="p-5 rounded-2xl bg-gradient-to-br from-slate-900/90 to-slate-900/50 border border-white/10 backdrop-blur-md shadow-xl relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
          <Activity size={64} className="text-purple-400" />
        </div>
        <div className="flex items-center space-x-3 text-purple-400 mb-2">
          <div className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/20">
            <Activity size={20} />
          </div>
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-400">Total Executions</span>
        </div>
        <div className="text-3xl font-extrabold text-white">{runCount}</div>
        <div className="text-xs text-gray-400 mt-1">Observed workflow runs</div>
      </div>

      {/* System Health */}
      <div className="p-5 rounded-2xl bg-gradient-to-br from-slate-900/90 to-slate-900/50 border border-white/10 backdrop-blur-md shadow-xl relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
          <CheckCircle2 size={64} className="text-emerald-400" />
        </div>
        <div className="flex items-center space-x-3 text-emerald-400 mb-2">
          <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20">
            <CheckCircle2 size={20} />
          </div>
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-400">Engine Health</span>
        </div>
        <div className="text-3xl font-extrabold text-emerald-400">99.9%</div>
        <div className="text-xs text-gray-400 mt-1">Go worker pool active</div>
      </div>
    </div>
  );
};
