import React, { useState } from 'react';
import { Search, Play, Edit3, Webhook, Clock, Trash2, ArrowUpRight, Sparkles, CheckCircle, AlertCircle } from 'lucide-react';
import { Workflow } from '../../types';

interface WorkflowListProps {
  workflows: Workflow[];
  onSelect: (workflow: Workflow) => void;
  onRun: (id: string) => void;
  onOpenWebhook: (workflow: Workflow) => void;
  onOpenSchedule: (workflow: Workflow) => void;
  onNewWorkflow: () => void;
}

export const WorkflowList: React.FC<WorkflowListProps> = ({
  workflows,
  onSelect,
  onRun,
  onOpenWebhook,
  onOpenSchedule,
  onNewWorkflow,
}) => {
  const [searchTerm, setSearchTerm] = useState('');

  const filteredWorkflows = workflows.filter(
    (w) =>
      w.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (w.description && w.description.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  return (
    <div className="space-y-6">
      {/* Search & Filter Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="relative w-full sm:w-80">
          <Search className="absolute left-3.5 top-3 text-gray-400" size={18} />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search workflows..."
            className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500 transition-colors"
          />
        </div>

        <div className="text-xs text-gray-400 font-mono">
          Showing {filteredWorkflows.length} of {workflows.length} workflows
        </div>
      </div>

      {/* Empty State */}
      {filteredWorkflows.length === 0 && (
        <div className="p-12 text-center rounded-2xl border border-dashed border-white/10 bg-slate-900/40 backdrop-blur-md">
          <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 text-indigo-400 flex items-center justify-center mx-auto mb-4">
            <Sparkles size={24} />
          </div>
          <h3 className="text-base font-semibold text-white">No workflows found</h3>
          <p className="text-sm text-gray-400 mt-1 max-w-md mx-auto">
            {searchTerm ? 'Try adjusting your search query or clear filters.' : 'Get started by creating your first visual automation graph.'}
          </p>
          {!searchTerm && (
            <button
              onClick={onNewWorkflow}
              className="mt-5 inline-flex items-center space-x-2 px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold shadow-lg shadow-indigo-600/30 transition-all hover:scale-105"
            >
              <Sparkles size={16} />
              <span>Create Workflow</span>
            </button>
          )}
        </div>
      )}

      {/* Workflow Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredWorkflows.map((workflow) => (
          <div
            key={workflow.id}
            className="rounded-2xl border border-white/10 bg-slate-900/70 hover:bg-slate-900/90 backdrop-blur-xl p-5 flex flex-col justify-between space-y-5 transition-all duration-300 hover:border-indigo-500/50 hover:shadow-2xl hover:shadow-indigo-500/10 group"
          >
            {/* Card Header */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <span
                  className={`inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-[11px] font-semibold border ${
                    workflow.is_active
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                      : 'bg-gray-500/10 text-gray-400 border-gray-500/20'
                  }`}
                >
                  {workflow.is_active ? <CheckCircle size={12} /> : <AlertCircle size={12} />}
                  <span>{workflow.is_active ? 'Active' : 'Inactive'}</span>
                </span>

                <span className="text-[10px] text-gray-400 font-mono">
                  {new Date(workflow.created_at).toLocaleDateString()}
                </span>
              </div>

              <h3
                onClick={() => onSelect(workflow)}
                className="text-lg font-bold text-white group-hover:text-indigo-300 transition-colors cursor-pointer flex items-center justify-between"
              >
                <span>{workflow.name}</span>
                <ArrowUpRight size={18} className="opacity-0 group-hover:opacity-100 transition-opacity text-indigo-400" />
              </h3>

              <p className="text-xs text-gray-400 mt-1.5 line-clamp-2">
                {workflow.description || 'No description provided.'}
              </p>
            </div>

            {/* Quick Action Buttons */}
            <div className="pt-4 border-t border-white/5 flex items-center justify-between">
              <div className="flex items-center space-x-1">
                <button
                  onClick={() => onOpenWebhook(workflow)}
                  className="p-2 rounded-lg text-purple-400 hover:bg-purple-500/10 hover:text-purple-300 transition-colors"
                  title="Webhook URL"
                >
                  <Webhook size={16} />
                </button>
                <button
                  onClick={() => onOpenSchedule(workflow)}
                  className="p-2 rounded-lg text-amber-400 hover:bg-amber-500/10 hover:text-amber-300 transition-colors"
                  title="Cron Schedule"
                >
                  <Clock size={16} />
                </button>
                <button
                  onClick={() => onSelect(workflow)}
                  className="p-2 rounded-lg text-blue-400 hover:bg-blue-500/10 hover:text-blue-300 transition-colors"
                  title="Edit Visual Canvas"
                >
                  <Edit3 size={16} />
                </button>
              </div>

              <button
                onClick={() => onRun(workflow.id)}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-xl bg-indigo-600/80 hover:bg-indigo-500 text-white text-xs font-semibold shadow transition-all hover:scale-105 active:scale-95"
              >
                <Play size={14} />
                <span>Run</span>
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
