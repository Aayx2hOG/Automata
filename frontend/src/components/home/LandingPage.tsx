import { ArrowRight, Clock, GitBranch, Plus, Webhook } from 'lucide-react';
import { Workflow, WorkflowRun } from '../../types';

interface LandingPageProps {
  workflows: Workflow[];
  runs: WorkflowRun[];
  loading?: boolean;
  onLaunchStudio: () => void;
  onOpenWorkflow: (workflow: Workflow) => void;
  onNavigateTab: (tab: 'workflows' | 'runs' | 'schedules' | 'webhooks') => void;
}

const statusColor = {
  pending: 'text-amber-300 bg-amber-400/10',
  running: 'text-sky-300 bg-sky-400/10',
  succeeded: 'text-emerald-300 bg-emerald-400/10',
  failed: 'text-rose-300 bg-rose-400/10',
};

export function LandingPage({ workflows, runs, loading, onLaunchStudio, onOpenWorkflow, onNavigateTab }: LandingPageProps) {
  const recentWorkflows = [...workflows].sort((a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at)).slice(0, 4);
  const recentRuns = [...runs].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)).slice(0, 5);
  const failedRuns = runs.filter(run => run.status === 'failed').length;

  return (
    <div className="space-y-8">
      <section className="flex flex-wrap items-end justify-between gap-6 pb-2">
        <div className="max-w-2xl space-y-3">
          <p className="text-xs uppercase tracking-widest text-[#ff9c8e] font-semibold">AUTOMATION WORKSPACE</p>
          <h1 className="text-3xl font-semibold tracking-tight">Workspace overview</h1>
          <p className="text-gray-400">Pick up a workflow, check its latest run, or build something new.</p>
        </div>
        <button onClick={onLaunchStudio} className="inline-flex items-center gap-2 rounded-xl bg-[#ff6d5a] px-5 py-3 text-sm font-semibold text-white hover:bg-[#ea4e43]">
          <Plus size={18} /> Create Workflow
        </button>
      </section>

      <section aria-label="Workspace summary" className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[
          { label: 'Saved workflows', value: workflows.length, detail: 'Ready to edit and run' },
          { label: 'Recent executions', value: runs.length, detail: 'Latest 50 runs across your workflows' },
          { label: 'Failed in recent runs', value: failedRuns, detail: 'Open execution history to investigate' },
        ].map(metric => (
          <div key={metric.label} className="rounded-xl border border-white/10 bg-[#19191c] p-5">
            <p className="text-sm text-gray-400">{metric.label}</p>
            <p className="my-2 text-3xl font-semibold tabular-nums">{loading ? '—' : metric.value}</p>
            <p className="text-xs text-gray-500">{metric.detail}</p>
          </div>
        ))}
      </section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <section aria-labelledby="recent-workflows" className="rounded-xl border border-white/10 bg-[#121214] p-6">
          <div className="mb-5 flex items-center justify-between gap-3">
            <h2 id="recent-workflows" className="font-semibold">Recent workflows</h2>
            <button onClick={() => onNavigateTab('workflows')} className="text-xs text-[#ff9c8e]">View all workflows →</button>
          </div>
          {loading ? <p className="text-sm text-gray-400">Loading workflows…</p> : recentWorkflows.length ? (
            <div className="space-y-2">
              {recentWorkflows.map(workflow => (
                <button key={workflow.id} onClick={() => onOpenWorkflow(workflow)} className="flex w-full items-center gap-4 rounded-xl border border-white/5 bg-[#19191c] p-4 text-left hover:border-[#ff6d5a]/50">
                  <GitBranch size={20} className="shrink-0 text-[#ff9c8e]" />
                  <div className="min-w-0 flex-1">
                    <h3 className="truncate text-sm font-semibold">{workflow.name}</h3>
                    <p className="mt-1 truncate text-xs text-gray-400">{workflow.description || 'Open the canvas to continue editing'}</p>
                    <p className="mt-2 text-xs text-gray-500">Updated {new Date(workflow.updated_at).toLocaleDateString()}</p>
                  </div>
                  <ArrowRight size={16} className="shrink-0 text-gray-400" />
                </button>
              ))}
            </div>
          ) : (
            <div className="rounded-xl border border-dashed border-white/15 p-7 text-center">
              <GitBranch className="mx-auto mb-3 text-[#ff9c8e]" />
              <h3 className="font-medium">Your first workflow starts here</h3>
              <p className="mt-2 text-sm text-gray-400">Choose a starter template or a blank canvas. Save it, run it, then inspect each step.</p>
              <button onClick={onLaunchStudio} className="mt-5 text-sm font-semibold text-[#ff9c8e]">Choose a template →</button>
            </div>
          )}
        </section>

        <section aria-labelledby="recent-runs" className="rounded-xl border border-white/10 bg-[#121214] p-6">
          <div className="mb-5 flex items-center justify-between gap-3">
            <h2 id="recent-runs" className="font-semibold">Recent executions</h2>
            <button onClick={() => onNavigateTab('runs')} className="text-xs text-[#ff9c8e]">View execution history →</button>
          </div>
          {loading ? <p className="text-sm text-gray-400">Loading executions…</p> : recentRuns.length ? (
            <ul className="divide-y divide-white/5">
              {recentRuns.map(run => (
                <li key={run.id} className="flex items-center justify-between gap-4 py-4">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{workflows.find(workflow => workflow.id === run.workflow_id)?.name || 'Workflow'}</p>
                    <p className="mt-1 text-xs text-gray-400">{run.trigger_type} · {new Date(run.created_at).toLocaleString()}</p>
                    {run.error_message && <p className="mt-1 line-clamp-2 text-xs text-rose-300">{run.error_message}</p>}
                  </div>
                  <span className={`rounded-full px-3 py-1 text-xs capitalize ${statusColor[run.status]}`}>{run.status}</span>
                </li>
              ))}
            </ul>
          ) : <p className="rounded-xl border border-dashed border-white/15 p-7 text-sm text-gray-400">No executions yet. Open a workflow and select Execute Workflow to see its results here.</p>}
        </section>
      </div>

      <section aria-label="Automation triggers" className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <button onClick={() => onNavigateTab('schedules')} className="flex items-center gap-4 rounded-xl border border-white/10 p-5 text-left hover:bg-white/5">
          <Clock className="shrink-0 text-amber-300" size={22} />
          <div><h2 className="text-sm font-semibold">Run on a schedule</h2><p className="mt-1 text-xs text-gray-400">Set a cron schedule for a saved workflow.</p></div>
          <ArrowRight className="ml-auto shrink-0 text-gray-500" size={16} />
        </button>
        <button onClick={() => onNavigateTab('webhooks')} className="flex items-center gap-4 rounded-xl border border-white/10 p-5 text-left hover:bg-white/5">
          <Webhook className="shrink-0 text-sky-300" size={22} />
          <div><h2 className="text-sm font-semibold">Connect a webhook</h2><p className="mt-1 text-xs text-gray-400">Send a signed payload and inspect the result.</p></div>
          <ArrowRight className="ml-auto shrink-0 text-gray-500" size={16} />
        </button>
      </section>
      <p className="text-xs text-gray-500">Visual workflows · Checkpointed execution · Configurable retries</p>
    </div>
  );
}
