export type RunStatus = 'pending' | 'running' | 'succeeded' | 'failed';
export type TriggerType = 'manual' | 'webhook' | 'cron' | 'interval' | 'api';

export interface GraphNode {
  label?: string;
  id: string;
  type: string;
  config: Record<string, any>;
  position?: { x: number; y: number };
}

export interface GraphEdge {
  from_node_id: string;
  to_node_id: string;
  condition?: string;
}

export interface WorkflowGraph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface Workflow {
  webhook_secret: string;
  id: string;
  owner_id: string;
  name: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface WorkflowVersion {
  id: string;
  workflow_id: string;
  version: number;
  graph: WorkflowGraph;
  created_at: string;
}

export interface WorkflowWithDetails extends Workflow {
  active_version?: WorkflowVersion;
  latest_run?: WorkflowRun;
  schedule_count?: number;
}

export interface WorkflowRun {
  id: string;
  workflow_id: string;
  workflow_version_id: string;
  status: RunStatus;
  trigger_type: TriggerType;
  outputs?: Record<string, any>;
  error_message?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
}

export interface Schedule {
  id: string;
  workflow_id: string;
  cron_expression: string;
  is_active: boolean;
  next_run_at: string;
  last_run_at?: string;
  created_at: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

export interface ValidationError {
  field?: string;
  message: string;
  node_id?: string;
}
