import React, { useState } from 'react';
import { PlayCircle, CheckCircle2, XCircle, Clock, AlertCircle, RefreshCw, Code2, ChevronRight } from 'lucide-react';
import { WorkflowRun } from '../../types';

interface RunsListProps {
  runs: WorkflowRun[];
  onRefresh: () => void;
}

export const RunsList: React.FC<RunsListProps> = ({ runs, onRefresh }) => {
  const [selectedRun, setSelectedRun] = useState<WorkflowRun | null>(null);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'succeeded':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 size={14} />
            <span>Succeeded</span>
          </span>
        );
      case 'failed':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-xs font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20">
            <XCircle size={14} />
            <span>Failed</span>
          </span>
        );
      case 'running':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-xs font-bold bg-[#ff6d5a]/10 text-[#ff6d5a] border border-[#ff6d5a]/20">
            <RefreshCw size={14} className="animate-spin" />
            <span>Running</span>
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-xs font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <AlertCircle size={14} />
            <span>Pending</span>
          </span>
        );
    }
  };

  return (
    <div className="space-y-6 font-sans">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-bold text-white flex items-center space-x-2">
            <PlayCircle className="text-[#ff6d5a]" size={22} />
            <span>Execution History</span>
          </h2>
          <p className="text-xs text-gray-400 mt-1">Real-time log of all triggered workflow runs and outputs.</p>
        </div>

        <button
          onClick={onRefresh}
          className="flex items-center space-x-2 px-3.5 py-2 rounded-xl bg-[#161a23] border border-white/10 hover:bg-[#1a202c] text-xs font-semibold text-gray-300 transition-colors"
        >
          <RefreshCw size={14} />
          <span>Refresh Logs</span>
        </button>
      </div>

      {/* Empty State */}
      {runs.length === 0 && (
        <div className="p-12 text-center rounded-2xl border border-white/10 bg-[#161a23]/60">
          <Clock className="w-10 h-10 text-gray-500 mx-auto mb-3" />
          <h3 className="text-sm font-bold text-white">No execution runs recorded yet</h3>
          <p className="text-xs text-gray-400 mt-1">Trigger a workflow manually or send a webhook payload to see runs.</p>
        </div>
      )}

      {/* Runs Table */}
      {runs.length > 0 && (
        <div className="rounded-2xl border border-white/10 bg-[#161a23] overflow-hidden shadow-2xl">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-[#0d1017] text-gray-400 uppercase tracking-wider font-bold border-b border-white/10">
                <tr>
                  <th className="px-5 py-3.5">Run ID</th>
                  <th className="px-5 py-3.5">Workflow ID</th>
                  <th className="px-5 py-3.5">Status</th>
                  <th className="px-5 py-3.5">Trigger Type</th>
                  <th className="px-5 py-3.5">Created At</th>
                  <th className="px-5 py-3.5 text-right">Details</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5 text-gray-200 font-mono">
                {runs.map((run) => (
                  <tr key={run.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-5 py-4 font-bold text-[#ff6d5a]">{run.id.slice(0, 8)}...</td>
                    <td className="px-5 py-4 text-gray-400">{run.workflow_id.slice(0, 8)}...</td>
                    <td className="px-5 py-4">{getStatusBadge(run.status)}</td>
                    <td className="px-5 py-4">
                      <span className="px-2 py-0.5 rounded bg-white/5 text-gray-300 uppercase text-[10px] font-sans font-bold">
                        {run.trigger_type}
                      </span>
                    </td>
                    <td className="px-5 py-4 text-gray-400">{new Date(run.created_at).toLocaleString()}</td>
                    <td className="px-5 py-4 text-right">
                      <button
                        onClick={() => setSelectedRun(run)}
                        className="inline-flex items-center space-x-1 text-[#ff6d5a] hover:text-[#ff8575] font-sans font-bold"
                      >
                        <span>View Outputs</span>
                        <ChevronRight size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Selected Run Details Drawer */}
      {selectedRun && (
        <div className="fixed inset-0 bg-[#0b0e14]/80 backdrop-blur-md z-50 flex items-center justify-center p-4">
          <div className="bg-[#161a23] border border-white/10 rounded-2xl max-w-3xl w-full p-6 shadow-2xl space-y-6">
            <div className="flex items-center justify-between border-b border-white/10 pb-4">
              <div>
                <h3 className="text-base font-bold text-white flex items-center space-x-2">
                  <Code2 className="text-[#ff6d5a]" size={18} />
                  <span>Execution Step Output Inspector</span>
                </h3>
                <p className="text-xs text-gray-400 font-mono mt-0.5">ID: {selectedRun.id}</p>
              </div>
              <button
                onClick={() => setSelectedRun(null)}
                className="text-xs px-3 py-1.5 rounded-lg bg-white/10 text-white hover:bg-white/20 font-semibold"
              >
                Close
              </button>
            </div>

            {/* Status */}
            <div className="flex items-center justify-between p-4 rounded-xl bg-[#0d1017] border border-white/5">
              <div>
                <span className="text-xs text-gray-400 uppercase font-bold">Trigger: {selectedRun.trigger_type}</span>
                <div className="mt-1">{getStatusBadge(selectedRun.status)}</div>
              </div>
              <div className="text-right text-xs text-gray-400 font-mono">
                <div>Started: {selectedRun.started_at ? new Date(selectedRun.started_at).toLocaleTimeString() : 'N/A'}</div>
                <div>Finished: {selectedRun.finished_at ? new Date(selectedRun.finished_at).toLocaleTimeString() : 'N/A'}</div>
              </div>
            </div>

            {selectedRun.error_message && (
              <div className="p-3 rounded-xl bg-rose-950/50 border border-rose-500/30 text-rose-300 text-xs font-mono">
                <strong>Error traceback:</strong> {selectedRun.error_message}
              </div>
            )}

            {/* Node Outputs */}
            <div className="space-y-2">
              <h4 className="text-xs font-bold text-gray-300 uppercase tracking-wider">Node Step Outputs (JSON)</h4>
              <div className="p-4 rounded-xl bg-[#0d1017] border border-white/10 text-emerald-400 font-mono text-xs max-h-64 overflow-y-auto">
                {selectedRun.outputs ? (
                  <pre>{JSON.stringify(selectedRun.outputs, null, 2)}</pre>
                ) : (
                  <span className="text-gray-500 italic">No node output object recorded.</span>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
