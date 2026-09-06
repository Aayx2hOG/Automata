import React, { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import {
  Webhook,
  Clock,
  Play,
  Globe,
  Timer,
  Terminal,
  GitFork,
  Code2,
  CheckCircle2,
  XCircle,
  AlertCircle,
} from 'lucide-react';

export interface CustomNodeData {
  label?: string;
  type: string;
  config?: Record<string, any>;
  status?: 'pending' | 'running' | 'succeeded' | 'failed';
  output?: any;
  error?: string;
}

const nodeMeta: Record<string, { title: string; color: string; bg: string; border: string; icon: React.FC<any>; isTrigger?: boolean }> = {
  webhook: {
    title: 'Webhook Trigger',
    color: 'text-purple-400',
    bg: 'bg-purple-950/40',
    border: 'border-purple-500/40',
    icon: Webhook,
    isTrigger: true,
  },
  cron: {
    title: 'Cron Schedule',
    color: 'text-amber-400',
    bg: 'bg-amber-950/40',
    border: 'border-amber-500/40',
    icon: Clock,
    isTrigger: true,
  },
  schedule: {
    title: 'Cron Schedule',
    color: 'text-amber-400',
    bg: 'bg-amber-950/40',
    border: 'border-amber-500/40',
    icon: Clock,
    isTrigger: true,
  },
  manual: {
    title: 'Manual Trigger',
    color: 'text-emerald-400',
    bg: 'bg-emerald-950/40',
    border: 'border-emerald-500/40',
    icon: Play,
    isTrigger: true,
  },
  http_request: {
    title: 'HTTP Request',
    color: 'text-blue-400',
    bg: 'bg-blue-950/40',
    border: 'border-blue-500/40',
    icon: Globe,
  },
  delay: {
    title: 'Delay Pause',
    color: 'text-pink-400',
    bg: 'bg-pink-950/40',
    border: 'border-pink-500/40',
    icon: Timer,
  },
  logger: {
    title: 'Logger',
    color: 'text-cyan-400',
    bg: 'bg-cyan-950/40',
    border: 'border-cyan-500/40',
    icon: Terminal,
  },
  condition: {
    title: 'Condition Logic',
    color: 'text-yellow-400',
    bg: 'bg-yellow-950/40',
    border: 'border-yellow-500/40',
    icon: GitFork,
  },
  json_parser: {
    title: 'JSON Parser',
    color: 'text-teal-400',
    bg: 'bg-teal-950/40',
    border: 'border-teal-500/40',
    icon: Code2,
  },
};

export const CustomNode = memo(({ data, selected }: NodeProps) => {
  const nodeData = data as unknown as CustomNodeData;
  const nodeType = nodeData.type || 'http_request';
  const meta = nodeMeta[nodeType] || {
    title: nodeType,
    color: 'text-indigo-400',
    bg: 'bg-indigo-950/40',
    border: 'border-indigo-500/40',
    icon: Code2,
  };
  const Icon = meta.icon;

  const isCondition = nodeType === 'condition';

  // Config display preview
  const getConfigSummary = () => {
    const config = nodeData.config || {};
    if (nodeType === 'http_request') {
      return `${config.method || 'GET'} ${config.url || 'https://...'}`;
    }
    if (nodeType === 'delay') {
      return `${config.duration_ms || config.delay_ms || 1000}ms pause`;
    }
    if (nodeType === 'cron' || nodeType === 'schedule') {
      return `Schedule: ${config.cron || '*/5 * * * *'}`;
    }
    if (nodeType === 'webhook') {
      return `Endpoint: POST /webhook/{id}`;
    }
    if (nodeType === 'condition') {
      return `${config.field || 'input'} ${config.operator || '=='} ${config.value || 'true'}`;
    }
    if (nodeType === 'logger') {
      return config.message ? `"${config.message}"` : 'Log context';
    }
    if (nodeType === 'json_parser') {
      return 'Parse input JSON';
    }
    return nodeData.label || meta.title;
  };

  return (
    <div
      className={`min-w-[240px] max-w-[320px] rounded-xl border ${meta.border} ${meta.bg} backdrop-blur-md shadow-xl transition-all duration-200 ${
        selected ? 'ring-2 ring-indigo-500 shadow-indigo-500/20' : 'hover:border-opacity-80'
      }`}
    >
      {/* Target handle for non-trigger nodes */}
      {!meta.isTrigger && (
        <Handle
          type="target"
          position={Position.Top}
          id="target"
          className="!bg-indigo-400 !w-3 !h-3 !-top-1.5 border-2 border-slate-900"
        />
      )}

      {/* Node Header */}
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-white/5">
        <div className="flex items-center space-x-2">
          <div className={`p-1.5 rounded-lg bg-white/5 ${meta.color}`}>
            <Icon size={16} />
          </div>
          <div>
            <h4 className="text-xs font-semibold text-gray-200 uppercase tracking-wider">
              {meta.isTrigger ? 'Trigger' : 'Action'}
            </h4>
            <p className="text-sm font-medium text-white">{nodeData.label || meta.title}</p>
          </div>
        </div>

        {/* Execution status badge if available */}
        {nodeData.status && (
          <div className="flex items-center">
            {nodeData.status === 'succeeded' && <CheckCircle2 size={16} className="text-emerald-400" />}
            {nodeData.status === 'failed' && <XCircle size={16} className="text-rose-400" />}
            {nodeData.status === 'running' && (
              <div className="w-4 h-4 border-2 border-indigo-400 border-t-transparent rounded-full animate-spin" />
            )}
            {nodeData.status === 'pending' && <AlertCircle size={16} className="text-amber-400" />}
          </div>
        )}
      </div>

      {/* Node Body / Summary */}
      <div className="p-3 text-xs text-gray-300 font-mono bg-black/20 rounded-b-xl break-all line-clamp-2">
        {getConfigSummary()}
      </div>

      {/* Condition handles vs standard output handle */}
      {isCondition ? (
        <div className="flex justify-between items-center px-4 py-1.5 text-[10px] font-semibold text-gray-400 border-t border-white/5">
          <div className="relative flex items-center space-x-1 text-emerald-400">
            <span>TRUE</span>
            <Handle
              type="source"
              position={Position.Bottom}
              id="true"
              style={{ left: '25%' }}
              className="!bg-emerald-400 !w-3 !h-3 !-bottom-1.5 border-2 border-slate-900"
            />
          </div>
          <div className="relative flex items-center space-x-1 text-rose-400">
            <span>FALSE</span>
            <Handle
              type="source"
              position={Position.Bottom}
              id="false"
              style={{ left: '75%' }}
              className="!bg-rose-400 !w-3 !h-3 !-bottom-1.5 border-2 border-slate-900"
            />
          </div>
        </div>
      ) : (
        <Handle
          type="source"
          position={Position.Bottom}
          id="source"
          className="!bg-indigo-400 !w-3 !h-3 !-bottom-1.5 border-2 border-slate-900"
        />
      )}
    </div>
  );
});
