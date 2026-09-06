import React, { useState } from 'react';
import { Search, Play, Edit3, Webhook, Clock, ArrowUpRight, Plus, CheckCircle, AlertCircle } from 'lucide-react';
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
      {/* Search Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="relative w-full sm:w-80">
          <Search className="absolute left-3.5 top-3 text-gray-400" size={18} />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Filter workflows..."
            className="w-full bg-[#161a23] border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
          />
        </div>

        <div className="text-xs text-gray-400 font-mono">
          Showing {filteredWorkflows.length} of {workflows.length} workflows
        </div>
      </div>

      {/* Empty State */}
      {filteredWorkflows.length === 0 && (
        <div className="p-12 text-center rounded-2xl border border-dashed border-white/10 bg-[#161a23]/60">
          <h3 className="text-base font-bold text-white">No workflows found</h3>
          <p className="text-xs text-gray-400 mt-1 max-w-md mx-auto">
            {searchTerm ? 'Try adjusting your search query.' : 'Create your first n8n-style visual workflow.'}
          </p>
          {!searchTerm && (
            <button
              onClick={onNewWorkflow}
              className="mt-5 inline-flex items-center space-x-2 px-4 py-2.5 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-xs font-bold shadow-lg shadow-[#ff6d5a]/30 transition-all hover:scale-105"
            >
              <Plus size={16} />
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
            className="rounded-2xl border border-white/10 bg-[#161a23] hover:border-[#ff6d5a]/60 p-5 flex flex-col justify-between space-y-5 transition-all duration-300 shadow-xl group"
          >
            {/* Card Header */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <span
                  className={`inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-[11px] font-bold border ${
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
                className="text-base font-bold text-white group-hover:text-[#ff6d5a] transition-colors cursor-pointer flex items-center justify-between"
              >
                <span>{workflow.name}</span>
                <ArrowUpRight size={16} className="opacity-0 group-hover:opacity-100 transition-opacity text-[#ff6d5a]" />
              </h3>

              <p className="text-xs text-gray-400 mt-1.5 line-clamp-2">
                {workflow.description || 'No description provided.'}
              </p>
            </div>

            {/* Quick Actions */}
            <div className="pt-4 border-t border-white/5 flex items-center justify-between">
              <div className="flex items-center space-x-1">
                <button
                  onClick={() => onOpenWebhook(workflow)}
                  className="p-2 rounded-lg text-purple-400 hover:bg-purple-500/10 transition-colors"
                  title="Webhook URL"
                >
                  <Webhook size={16} />
                </button>
                <button
                  onClick={() => onOpenSchedule(workflow)}
                  className="p-2 rounded-lg text-amber-400 hover:bg-amber-500/10 transition-colors"
                  title="Cron Schedule"
                >
                  <Clock size={16} />
                </button>
                <button
                  onClick={() => onSelect(workflow)}
                  className="p-2 rounded-lg text-cyan-400 hover:bg-cyan-500/10 transition-colors"
                  title="Edit Canvas"
                >
                  <Edit3 size={16} />
                </button>
              </div>

              <button
                onClick={() => onRun(workflow.id)}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-xs font-bold shadow-md shadow-[#ff6d5a]/20 transition-all hover:scale-105"
              >
                <Play size={14} />
                <span>Execute</span>
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
