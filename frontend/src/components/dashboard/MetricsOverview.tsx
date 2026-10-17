import { Layers, Clock, Activity, CheckCircle2 } from 'lucide-react';
import type { Workflow } from '../../types';

interface MetricsOverviewProps {
  workflows: Workflow[];
  runCount?: number;
  scheduleCount?: number;
  loading?: boolean;
}
export function MetricsOverview({ workflows, runCount = 0, scheduleCount = 0, loading }: MetricsOverviewProps) {
  const metrics = [
    { label: 'Total workflows', value: workflows.length, icon: Layers },
    { label: 'Active workflows', value: workflows.filter(workflow => workflow.is_active).length, icon: CheckCircle2 },
    { label: 'Recent executions', value: runCount, icon: Activity },
    { label: 'Active schedules', value: scheduleCount, icon: Clock },
  ];
  return <section aria-label="Workspace metrics" className="workspace-metrics">{metrics.map(({ label, value, icon: Icon }) => <div key={label}><Icon size={17} /><span>{label}</span><strong>{loading ? '—' : value}</strong></div>)}</section>;
}
