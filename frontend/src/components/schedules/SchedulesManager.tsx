import React, { useMemo, useState } from 'react';
import { Clock, Plus, Trash2, Calendar } from 'lucide-react';
import { Schedule, Workflow, ScheduleInput, ScheduleOptions } from '../../types';

interface SchedulesManagerProps {
  workflows: Workflow[];
  initialWorkflowId?: string;
  schedules: Schedule[];
  onCreateSchedule: (workflowId: string, cronExpression: ScheduleInput) => Promise<void>;
  onDeactivateSchedule: (scheduleId: string) => Promise<void>;
}

export const SchedulesManager: React.FC<SchedulesManagerProps> = ({
  workflows,
  initialWorkflowId = '',
  schedules,
  onCreateSchedule,
  onDeactivateSchedule,
}) => {
  const [selectedId, setSelectedWorkflowId] = useState(initialWorkflowId);
  const selectedWorkflowId = workflows.find(w => w.id === selectedId)?.id || workflows[0]?.id || '';
  const [cronExpression, setCronExpression] = useState('*/5 * * * *');
  const [mode, setMode] = useState<'simple' | 'cron'>('simple');
  const [options, setOptions] = useState<ScheduleOptions>({ frequency: 'daily', time: '09:00', weekday: 1, day: 1, minute: 0, timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC' });
  const timeZones = useMemo(() => {
    const available = typeof Intl.supportedValuesOf === 'function' ? Intl.supportedValuesOf('timeZone') : ['America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles', 'Europe/London', 'Europe/Paris', 'Asia/Dubai', 'Asia/Kolkata', 'Asia/Singapore', 'Asia/Tokyo', 'Australia/Sydney'];
    return [...new Set(['UTC', options.timezone, 'Asia/Kolkata', ...available])].sort();
  }, [options.timezone]);
  const fieldClass = 'mt-1 w-full bg-[#101012] border border-white/10 rounded-xl px-3 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a]';
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const weekdays = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const [hour, minute] = options.time.split(':').map(Number);
  const displayTime = options.time ? new Intl.DateTimeFormat(undefined, { hour: 'numeric', minute: '2-digit' }).format(new Date(2000, 0, 1, hour, minute)) : 'your chosen time';
  const summary = options.frequency === 'hourly' ? `Every hour at :${String(options.minute).padStart(2, '0')}`
    : options.frequency === 'weekly' ? `Every ${weekdays[options.weekday]} at ${displayTime}`
    : options.frequency === 'monthly' ? `On day ${options.day} of each month at ${displayTime}`
    : `Every day at ${displayTime}`;


  const [error, setError] = useState<string | null>(null);
  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedWorkflowId || (mode === 'cron' && !cronExpression.trim())) return;

    setIsSubmitting(true); setError(null); setSuccess(false);
    try {
      await onCreateSchedule(selectedWorkflowId, mode === 'cron' ? cronExpression : options);
      setSuccess(true);
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
          <span>Schedule Manager</span>
        </h2>
        <p className="text-xs text-gray-400 mt-1">Choose when your workflow should run. We’ll take care of the rest.</p>
      </div>

      {/* Schedule Creation Card */}
      <div className="p-6 rounded-2xl bg-[#19191c] border border-white/10 shadow-xl space-y-4">
        <h3 className="text-sm font-bold text-white flex items-center space-x-2">
          <Plus size={16} className="text-[#ff6d5a]" />
          <span>Create New Schedule</span>
        </h3>

        <form onSubmit={handleCreate} className="space-y-5" onChange={() => setSuccess(false)}>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <label className="block text-xs font-medium text-gray-300">Workflow
              <select aria-label="Select workflow" className={fieldClass} value={selectedWorkflowId} disabled={!workflows.length} onChange={e => setSelectedWorkflowId(e.target.value)}>
                {!workflows.length && <option value="">Create a workflow first</option>}
                {workflows.map(w => <option key={w.id} value={w.id}>{w.name}</option>)}
              </select>
            </label>
            {mode === 'simple' && <>
              <label className="block text-xs font-medium text-gray-300">Time
                <input required type="time" step="60" style={{ colorScheme: 'dark' }} className={fieldClass} value={options.time} onChange={e => setOptions({ ...options, time: e.target.value, minute: Number(e.target.value.split(':')[1] || 0) })} />
              </label>
              <label className="block text-xs font-medium text-gray-300">Repeat
                <select className={fieldClass} value={options.frequency} onChange={e => setOptions({ ...options, frequency: e.target.value as ScheduleOptions['frequency'] })}>
                  <option value="daily">Every day</option><option value="weekly">Every week</option><option value="monthly">Every month</option><option value="hourly">Every hour</option>
                </select>
              </label>
              {options.frequency === 'weekly' && <label className="block text-xs font-medium text-gray-300">Day of week
                <select className={fieldClass} value={options.weekday} onChange={e => setOptions({ ...options, weekday: Number(e.target.value) })}>
                  {weekdays.map((day, index) => <option key={day} value={index}>{day}</option>)}
                </select>
              </label>}
              {options.frequency === 'monthly' && <label className="block text-xs font-medium text-gray-300">Day of month (1–28)
                <select className={fieldClass} value={options.day} onChange={e => setOptions({ ...options, day: Number(e.target.value) })}>
                  {Array.from({ length: 28 }, (_, index) => index + 1).map(day => <option key={day} value={day}>{day}</option>)}
                </select>
              </label>}
            </>}
          </div>
          {mode === 'simple' && <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-gray-400">
            <span>Times are in {options.timezone.replaceAll('_', ' ')}.</span>
            <details>
              <summary className="cursor-pointer text-[#ff9c8e]">Change time zone</summary>
              <label className="block mt-3">Time zone
                <select className={fieldClass} value={options.timezone} onChange={e => setOptions({ ...options, timezone: e.target.value })}>
                  {timeZones.map(zone => <option key={zone} value={zone}>{zone === 'UTC' ? 'UTC — Coordinated Universal Time' : zone.replaceAll('_', ' ').replaceAll('/', ' / ')}</option>)}
                </select>
              </label>
            </details>
          </div>}
          {mode === 'simple' && options.frequency === 'hourly' && <p className="text-xs text-gray-400">Uses the minutes from your selected time: 9:30 means 10:30, 11:30, and so on.</p>}
          <div>
            <button type="button" className="text-xs text-gray-400 hover:text-white" aria-expanded={mode === 'cron'} aria-controls="advanced-schedule" onClick={() => { setMode(mode === 'cron' ? 'simple' : 'cron'); setSuccess(false); }}>{mode === 'cron' ? '← Back to time picker' : 'Advanced: use a cron expression'}</button>
            {mode === 'cron' && <div id="advanced-schedule" className="mt-3 max-w-lg">
              <label className="block text-xs text-gray-300">Cron expression
                <input required className={fieldClass} value={cronExpression} onChange={e => setCronExpression(e.target.value)} placeholder="*/5 * * * *" />
              </label>
              <p className="mt-2 text-xs text-gray-400">Five fields: minute, hour, day, month, weekday. Default time zone: UTC.</p>
            </div>}
          </div>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-t border-white/10 pt-5">
            <p aria-live="polite" className="text-sm text-gray-200">{mode === 'simple' ? summary : 'Runs using your custom cron expression'}</p>
            <button type="submit" disabled={isSubmitting || !selectedWorkflowId} className="shrink-0 px-5 py-3 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-sm font-semibold disabled:opacity-50">
              {isSubmitting ? 'Creating…' : 'Create schedule'}
            </button>
          </div>
          {success && <p role="status" className="text-sm text-emerald-400">Schedule created. Your workflow will run automatically.</p>}
        </form>
      </div>

      {/* Active Schedules Table */}
      <div className="rounded-2xl border border-white/10 bg-[#19191c] overflow-hidden shadow-2xl">
        <div className="px-6 py-4 border-b border-white/10 font-bold text-sm text-white flex items-center space-x-2">
          <Calendar size={18} className="text-amber-400" />
          <span>Active Schedules ({schedules.length})</span>
        </div>

        {schedules.length === 0 ? (
          <div className="p-8 text-center text-xs text-gray-400">
            No active schedules configured. Use the form above to add a schedule.
          </div>
        ) : (
          <table className="w-full text-left text-xs">
            <thead className="bg-[#101012] text-gray-400 uppercase tracking-wider font-bold border-b border-white/10">
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
