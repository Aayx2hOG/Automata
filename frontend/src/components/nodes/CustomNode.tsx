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

const nodeMeta: Record<
  string,
  { title: string; color: string; bg: string; border: string; icon: React.FC<any>; isTrigger?: boolean }
> = {
  webhook_trigger: {
    title: 'Webhook Trigger',
    color: 'text-[#ff6d5a]',
    bg: 'bg-[#ff6d5a]/10',
    border: 'border-[#ff6d5a]/40',
    icon: Webhook,
    isTrigger: true,
  },
  cron: {
    title: 'Cron Schedule',
    color: 'text-[#ff6d5a]',
    bg: 'bg-[#ff6d5a]/10',
    border: 'border-[#ff6d5a]/40',
    icon: Clock,
    isTrigger: true,
  },
  schedule: {
    title: 'Cron Schedule',
    color: 'text-[#ff6d5a]',
    bg: 'bg-[#ff6d5a]/10',
    border: 'border-[#ff6d5a]/40',
    icon: Clock,
    isTrigger: true,
  },
  manual_trigger: {
    title: 'Manual Trigger',
    color: 'text-[#ff6d5a]',
    bg: 'bg-[#ff6d5a]/10',
    border: 'border-[#ff6d5a]/40',
    icon: Play,
    isTrigger: true,
  },
  http_request: {
    title: 'HTTP Request',
    color: 'text-cyan-400',
    bg: 'bg-cyan-950/40',
    border: 'border-cyan-500/40',
    icon: Globe,
  },
  delay: {
    title: 'Delay Pause',
    color: 'text-purple-400',
    bg: 'bg-purple-950/40',
    border: 'border-purple-500/40',
    icon: Timer,
  },
  logger: {
    title: 'Logger',
    color: 'text-blue-400',
    bg: 'bg-blue-950/40',
    border: 'border-blue-500/40',
    icon: Terminal,
  },
  condition: {
    title: 'Condition Logic',
    color: 'text-amber-400',
    bg: 'bg-amber-950/40',
    border: 'border-amber-500/40',
    icon: GitFork,
  },
  json_parser: {
    title: 'JSON Parser',
    color: 'text-emerald-400',
    bg: 'bg-emerald-950/40',
    border: 'border-emerald-500/40',
    icon: Code2,
  },
};

export const CustomNode = memo(({ data, selected }: NodeProps) => {
  const nodeData = data as unknown as CustomNodeData;
  const nodeType = nodeData.type || 'http_request';
  const meta = nodeMeta[nodeType] || {
    title: nodeType,
    color: 'text-cyan-400',
    bg: 'bg-cyan-950/40',
    border: 'border-cyan-500/40',
    icon: Code2,
  };
  const Icon = meta.icon;
  const isCondition = nodeType === 'condition';

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
    if (nodeType === 'webhook_trigger') {
      return `POST /webhook/{id}`;
    }
    if (nodeType === 'condition') {
      return `${config.field || 'input'} ${config.operator || '=='} ${config.value || 'true'}`;
    }
    if (nodeType === 'logger') {
      return config.message ? `"${config.message}"` : 'Log context';
    }
    if (nodeType === 'json_parser') {
      return 'Parse JSON';
    }
    return nodeData.label || meta.title;
  };

  return (
    <div
      className={`min-w-[240px] max-w-[320px] rounded-xl border bg-[#161a23] backdrop-blur-md shadow-2xl transition-all duration-200 ${
        selected ? 'ring-2 ring-[#ff6d5a] shadow-[#ff6d5a]/20 border-[#ff6d5a]' : meta.border
      }`}
    >
      {/* Target handle for non-trigger nodes */}
      {!meta.isTrigger && (
        <Handle
          type="target"
          position={Position.Top}
          id="target"
          className="!bg-[#ff6d5a] !w-3 !h-3 !-top-1.5 border-2 border-[#161a23]"
        />
      )}

      {/* n8n Style Node Header */}
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-white/5">
        <div className="flex items-center space-x-2.5">
          <div className={`p-1.5 rounded-lg ${meta.bg} ${meta.color} border ${meta.border}`}>
            <Icon size={16} />
          </div>
          <div>
            <span className="text-[10px] font-bold uppercase tracking-wider text-gray-400 block">
              {meta.isTrigger ? 'Trigger' : 'Node'}
            </span>
            <p className="text-xs font-bold text-white leading-none mt-0.5">{nodeData.label || meta.title}</p>
          </div>
        </div>

        {/* Execution status badge */}
        {nodeData.status && (
          <div className="flex items-center">
            {nodeData.status === 'succeeded' && <CheckCircle2 size={16} className="text-emerald-400" />}
            {nodeData.status === 'failed' && <XCircle size={16} className="text-rose-400" />}
            {nodeData.status === 'running' && (
              <div className="w-4 h-4 border-2 border-[#ff6d5a] border-t-transparent rounded-full animate-spin" />
            )}
            {nodeData.status === 'pending' && <AlertCircle size={16} className="text-amber-400" />}
          </div>
        )}
      </div>

      {/* Node Config Summary */}
      <div className="p-3 text-[11px] text-gray-300 font-mono bg-[#0d1017]/80 rounded-b-xl break-all line-clamp-2">
        {getConfigSummary()}
      </div>

      {/* Handles */}
      {isCondition ? (
        <div className="flex justify-between items-center px-4 py-1 text-[10px] font-bold text-gray-400 border-t border-white/5">
          <div className="relative flex items-center space-x-1 text-emerald-400">
            <span>TRUE</span>
            <Handle
              type="source"
              position={Position.Bottom}
              id="true"
              style={{ left: '25%' }}
              className="!bg-emerald-400 !w-3 !h-3 !-bottom-1.5 border-2 border-[#161a23]"
            />
          </div>
          <div className="relative flex items-center space-x-1 text-rose-400">
            <span>FALSE</span>
            <Handle
              type="source"
              position={Position.Bottom}
              id="false"
              style={{ left: '75%' }}
              className="!bg-rose-400 !w-3 !h-3 !-bottom-1.5 border-2 border-[#161a23]"
            />
          </div>
        </div>
      ) : (
        <Handle
          type="source"
          position={Position.Bottom}
          id="source"
          className="!bg-[#ff6d5a] !w-3 !h-3 !-bottom-1.5 border-2 border-[#161a23]"
        />
      )}
    </div>
  );
});
