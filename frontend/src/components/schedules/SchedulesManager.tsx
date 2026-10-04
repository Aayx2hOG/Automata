import React, { useState } from 'react';
import { Clock, Plus, Trash2, Calendar } from 'lucide-react';
import { Schedule, Workflow } from '../../types';

interface SchedulesManagerProps {
  workflows: Workflow[];
  schedules: Schedule[];
  onCreateSchedule: (workflowId: string, cronExpression: string) => Promise<void>;
  onDeactivateSchedule: (scheduleId: string) => Promise<void>;
}

export const SchedulesManager: React.FC<SchedulesManagerProps> = ({
  workflows,
  schedules,
  onCreateSchedule,
  onDeactivateSchedule,
}) => {
  const [selectedId, setSelectedWorkflowId] = useState('');
  const selectedWorkflowId = workflows.find(w => w.id === selectedId)?.id || workflows[0]?.id || '';
  const [cronExpression, setCronExpression] = useState('*/5 * * * *');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedWorkflowId || !cronExpression.trim()) return;

    setIsSubmitting(true); setError(null);
    try {
      await onCreateSchedule(selectedWorkflowId, cronExpression);
    } catch (err) {setError((err as Error).message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-8 font-sans">
      {error && <p role="alert" className="text-rose-300">{error}</p>}
      <div>
        <h2 className="text-lg font-bold text-white flex items-center space-x-2">
          <Clock className="text-amber-400" size={22} />
          <span>Cron Schedule Manager</span>
        </h2>
        <p className="text-xs text-gray-400 mt-1">Configure automated time-based execution schedules for workflows.</p>
      </div>

      {/* Schedule Creation Card */}
      <div className="p-6 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl space-y-4">
        <h3 className="text-sm font-bold text-white flex items-center space-x-2">
          <Plus size={16} className="text-[#ff6d5a]" />
          <span>Create New Schedule</span>
        </h3>

        <form onSubmit={handleCreate} className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
          <div>
            <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Select Workflow</label>
            {workflows.length === 0 ? (
              <div className="mt-1 rounded-xl border border-dashed border-white/15 bg-[#0d1017] px-3 py-3 text-sm text-gray-500">
                No workflows created yet. Create a workflow before adding a schedule.
              </div>
            ) : (
              <select
                value={selectedWorkflowId}
                onChange={(e) => setSelectedWorkflowId(e.target.value)}
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-xl px-3 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a]"
              >
                {workflows.map((wf) => (
                  <option key={wf.id} value={wf.id}>
                    {wf.name}
                  </option>
                ))}
              </select>
            )}
          </div>

          <div>
            <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Cron Expression</label>
            <input
              type="text"
              value={cronExpression}
              onChange={(e) => setCronExpression(e.target.value)}
              placeholder="*/5 * * * *"
              className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-xl px-3 py-2.5 text-sm text-amber-300 font-mono focus:outline-none focus:border-[#ff6d5a]"
            />
          </div>

          <button
            type="submit"
            disabled={isSubmitting || !selectedWorkflowId}
            className="px-5 py-2.5 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-xs font-bold shadow-lg shadow-[#ff6d5a]/30 transition-all hover:scale-105 disabled:opacity-50"
          >
            {isSubmitting ? 'Creating...' : 'Activate Schedule'}
          </button>
        </form>
      </div>

      {/* Active Schedules Table */}
      <div className="rounded-2xl border border-white/10 bg-[#161a23] overflow-hidden shadow-2xl">
        <div className="px-6 py-4 border-b border-white/10 font-bold text-sm text-white flex items-center space-x-2">
          <Calendar size={18} className="text-amber-400" />
          <span>Active Schedules ({schedules.length})</span>
        </div>

        {schedules.length === 0 ? (
          <div className="p-8 text-center text-xs text-gray-400">
            No active schedules configured. Use the form above to add a cron schedule.
          </div>
        ) : (
          <table className="w-full text-left text-xs">
            <thead className="bg-[#0d1017] text-gray-400 uppercase tracking-wider font-bold border-b border-white/10">
              <tr>
                <th className="px-6 py-3">Schedule ID</th>
                <th className="px-6 py-3">Workflow ID</th>
                <th className="px-6 py-3">Cron Expression</th>
                <th className="px-6 py-3">Next Execution</th>
                <th className="px-6 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 font-mono text-gray-200">
              {schedules.map((sched) => (
                <tr key={sched.id} className="hover:bg-white/5 transition-colors">
                  <td className="px-6 py-4 text-[#ff6d5a] font-bold">{sched.id.slice(0, 8)}...</td>
                  <td className="px-6 py-4 text-gray-400">{sched.workflow_id.slice(0, 8)}...</td>
                  <td className="px-6 py-4 font-bold text-amber-400">{sched.cron_expression}</td>
                  <td className="px-6 py-4 text-gray-300">
                    {sched.next_run_at ? new Date(sched.next_run_at).toLocaleString() : 'Pending'}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => onDeactivateSchedule(sched.id).catch(err => setError(err.message))}
                      className="p-1.5 rounded-lg text-rose-400 hover:bg-rose-500/10 transition-colors"
                      title="Deactivate Schedule"
                    >
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
