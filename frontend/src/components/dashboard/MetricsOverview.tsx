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
      <div className="p-5 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl relative overflow-hidden group">
        <div className="flex items-center space-x-3 text-[#ff6d5a] mb-2">
          <div className="p-2 rounded-xl bg-[#ff6d5a]/10 border border-[#ff6d5a]/20">
            <Layers size={18} />
          </div>
          <span className="text-xs font-bold uppercase tracking-wider text-gray-400">Total Workflows</span>
        </div>
        <div className="text-3xl font-extrabold text-white">{workflows.length}</div>
        <div className="text-xs text-gray-400 mt-1">{activeWorkflows} active DAG pipelines</div>
      </div>

      {/* Active Schedules */}
      <div className="p-5 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl relative overflow-hidden group">
        <div className="flex items-center space-x-3 text-amber-400 mb-2">
          <div className="p-2 rounded-xl bg-amber-500/10 border border-amber-500/20">
            <Clock size={18} />
          </div>
          <span className="text-xs font-bold uppercase tracking-wider text-gray-400">Cron Schedules</span>
        </div>
        <div className="text-3xl font-extrabold text-white">{scheduleCount}</div>
        <div className="text-xs text-gray-400 mt-1">Automated interval triggers</div>
      </div>

      {/* Executions */}
      <div className="p-5 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl relative overflow-hidden group">
        <div className="flex items-center space-x-3 text-cyan-400 mb-2">
          <div className="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/20">
            <Activity size={18} />
          </div>
          <span className="text-xs font-bold uppercase tracking-wider text-gray-400">Executions</span>
        </div>
        <div className="text-3xl font-extrabold text-white">{runCount}</div>
        <div className="text-xs text-gray-400 mt-1">Observed workflow runs</div>
      </div>

      {/* System Health */}
      <div className="p-5 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl relative overflow-hidden group">
        <div className="flex items-center space-x-3 text-emerald-400 mb-2">
          <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20">
            <CheckCircle2 size={18} />
          </div>
          <span className="text-xs font-bold uppercase tracking-wider text-gray-400">Engine Status</span>
        </div>
        <div className="text-3xl font-extrabold text-emerald-400">Healthy</div>
        <div className="text-xs text-gray-400 mt-1">Go worker pool online</div>
      </div>
    </div>
  );
};
