import React, { useState, useRef } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { Trash2, Search, Play, Edit3, Webhook, Clock, ArrowUpRight, Plus, CheckCircle, AlertCircle, LayoutGrid, List, GitBranch } from 'lucide-react';
import { Workflow } from '../../types';

interface WorkflowListProps {
  workflows: Workflow[];
  onDelete: (workflow: Workflow) => Promise<void>;
  onSelect: (workflow: Workflow) => void;
  onRun: (id: string) => void;
  onOpenWebhook: (workflow: Workflow) => void;
  onOpenSchedule: (workflow: Workflow) => void;
  onNewWorkflow: () => void;
}

export const WorkflowList: React.FC<WorkflowListProps> = ({
  workflows,
  onDelete,
  onSelect,
  onRun,
  onOpenWebhook,
  onOpenSchedule,
  onNewWorkflow,
}) => {
  const [deleteTarget, setDeleteTarget] = useState<Workflow | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const returnFocus = useRef<HTMLElement | null>(null);
  const searchInput = useRef<HTMLInputElement | null>(null);
  const confirmDelete = async () => {
    if (!deleteTarget || deleting) return;
    setDeleting(true);
    setDeleteError(null);
    try {
      await onDelete(deleteTarget);
      setDeleteTarget(null);
    } catch (error) {
      setDeleteError(error instanceof Error && error.name === 'TimeoutError' ? 'The server took too long to respond. Refresh the page to check whether the workflow was deleted, then try again if needed.' : error instanceof Error && error.message.trim() ? error.message : 'Unable to delete workflow. Please try again.');
    } finally {
      setDeleting(false);
    }
  };
  const [searchTerm, setSearchTerm] = useState('');
  const [status, setStatus] = useState('all');
  const [sort, setSort] = useState('updated');
  const [view, setView] = useState<'grid' | 'list'>('grid');

  const filteredWorkflows = workflows.filter(w => status === 'all' || w.is_active === (status === 'active')).filter(
    (w) =>
      w.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (w.description && w.description.toLowerCase().includes(searchTerm.toLowerCase()))
  ).sort((a, b) => sort === 'name' ? a.name.localeCompare(b.name) : Date.parse(b.updated_at) - Date.parse(a.updated_at));

  return (
    <div className="space-y-6">
      {/* Search Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="relative w-full sm:w-80">
          <Search className="absolute left-3.5 top-3 text-gray-400" size={18} />
          <input
            ref={searchInput}
            aria-label="Filter workflows"
            type="search"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Filter workflows..."
            className="w-full bg-[#19191c] border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
          />
        </div>

        <div className="workflow-filters">
          <select aria-label="Workflow status" value={status} onChange={event => setStatus(event.target.value)}><option value="all">All statuses</option><option value="active">Active</option><option value="inactive">Inactive</option></select>
          <select aria-label="Sort workflows" value={sort} onChange={event => setSort(event.target.value)}><option value="updated">Last updated</option><option value="name">Name A–Z</option></select>
          <div className="view-switch" aria-label="Workflow view"><button aria-label="Grid view" aria-pressed={view === 'grid'} onClick={() => setView('grid')}><LayoutGrid size={16} /></button><button aria-label="List view" aria-pressed={view === 'list'} onClick={() => setView('list')}><List size={16} /></button></div>
        </div>
      </div>

      {/* Empty State */}
      {filteredWorkflows.length === 0 && (
        <div className="p-12 text-center rounded-2xl border border-dashed border-white/10 bg-[#19191c]/60">
          <h3 className="text-base font-bold text-white">No workflows found</h3>
          <p className="text-xs text-gray-400 mt-1 max-w-md mx-auto">
            {searchTerm || status !== 'all' ? 'Try another search or status filter.' : 'Create your first visual workflow.'}
          </p>
          {!searchTerm && status === 'all' && (
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

      <p className="text-xs text-gray-500">{filteredWorkflows.length} of {workflows.length} workflows</p>
      {/* Workflow Cards Grid */}
      <div className={`workflow-collection ${view === 'list' ? 'workflow-list-view' : 'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4'}`}>
        {filteredWorkflows.map((workflow) => (
          <div
            key={workflow.id}
            className="workflow-card rounded-xl border border-white/10 bg-[#161618] hover:border-[#ff6d5a]/40 p-5 flex flex-col justify-between gap-5 transition-colors group"
          >
            {/* Card Header */}
            <div className="workflow-card-info">
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
                  Updated {new Date(workflow.updated_at).toLocaleDateString()}
                </span>
              </div>

              <h3
                className="text-base font-bold text-white group-hover:text-[#ff6d5a] transition-colors cursor-pointer flex items-center justify-between"
              >
                <button onClick={() => onSelect(workflow)} className="flex w-full items-center justify-between text-left">
                  <span className="flex min-w-0 items-center gap-2"><GitBranch size={16} className="shrink-0 text-gray-500" /><span className="truncate">{workflow.name}</span></span>
                  <ArrowUpRight size={16} className="text-[#ff6d5a]" />
                </button>
              </h3>

              <p className="text-xs text-gray-400 mt-1.5 line-clamp-2">
                {workflow.description || 'No description provided.'}
              </p>
            </div>

            {/* Quick Actions */}
            <div className="workflow-card-actions pt-4 border-t border-white/5 flex items-center justify-between">
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
                <button
                  onClick={event => { returnFocus.current = event.currentTarget; setDeleteError(null); setDeleteTarget(workflow); }}
                  className="p-2 rounded-lg text-gray-400 hover:text-[#f29b83] hover:bg-[#f29b83]/10 transition-colors"
                  title="Delete workflow"
                  aria-label={`Delete ${workflow.name}`}
                >
                  <Trash2 size={16} />
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
      <Dialog.Root open={deleteTarget !== null} onOpenChange={open => { if (!open && !deleting) setDeleteTarget(null); }}>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content
            className="workflow-dialog delete-workflow-dialog"
            onCloseAutoFocus={event => { event.preventDefault(); (returnFocus.current?.isConnected ? returnFocus.current : searchInput.current)?.focus(); }}
            onEscapeKeyDown={event => { if (deleting) event.preventDefault(); }}
            onPointerDownOutside={event => event.preventDefault()}
          >
            <div className="delete-workflow-icon"><Trash2 size={22} /></div>
            <Dialog.Title className="text-lg font-semibold">Delete workflow?</Dialog.Title>
            <Dialog.Description className="mt-3 text-sm leading-6 text-gray-400">
              Permanently delete <strong className="break-words text-gray-200">{deleteTarget?.name}</strong>, its saved versions, schedules, and execution history? Its webhook will stop working. This cannot be undone.
            </Dialog.Description>
            {deleteError && <p role="alert" className="mt-4 rounded-lg border border-[#f29b83]/20 bg-[#f29b83]/5 p-3 text-sm text-[#f0b7a6]">{deleteError}</p>}
            <div className="mt-6 flex justify-end gap-3">
              <button type="button" disabled={deleting} onClick={() => setDeleteTarget(null)} className="rounded-lg border border-white/10 px-4 py-2 text-sm text-gray-300 disabled:opacity-50">Cancel</button>
              <button type="button" disabled={deleting} onClick={() => void confirmDelete()} className="delete-workflow-confirm"><Trash2 size={15} />{deleting ? 'Deleting…' : 'Delete workflow'}</button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
};
