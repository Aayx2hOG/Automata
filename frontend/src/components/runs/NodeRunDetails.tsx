import { NodeRun } from '../../types';

export function NodeRunDetails({ nodes = [] }: { nodes?: NodeRun[] }) {
  if (!nodes.length) return <p className="text-xs text-gray-400">No node checkpoints recorded yet.</p>;
  return <div className="max-h-64 space-y-2 overflow-auto" aria-label="Node executions">
    {nodes.map(node => {
      const duration = node.started_at ? Math.max(0, (new Date(node.updated_at).getTime() - new Date(node.started_at).getTime()) / 1000) : null;
      return <details key={node.node_id} className="rounded-lg border border-white/10 bg-[#0d1017] p-3 text-xs">
        <summary className="cursor-pointer break-all text-gray-200">
          {node.node_id} · <span className={node.status === 'failed' ? 'text-rose-300' : node.status === 'succeeded' ? 'text-emerald-300' : 'text-amber-300'}>{node.status}</span>
          {' · '}{node.attempts} {node.attempts === 1 ? 'attempt' : 'attempts'}
          {duration !== null && node.status !== 'running' && ` · ${duration.toFixed(2)}s elapsed`}
        </summary>
        {node.error_message && <p className="mt-2 text-rose-300">{node.error_message}</p>}
        {node.output && <pre className="mt-2 overflow-auto text-emerald-300">{JSON.stringify(node.output, null, 2)}</pre>}
      </details>;
    })}
  </div>;
}
