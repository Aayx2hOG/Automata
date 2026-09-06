import React, { useEffect, useState } from 'react';
import { X, Trash2, Sliders, Play, Code2, Globe, Clock, Terminal, GitFork, Timer, Webhook } from 'lucide-react';
import { CustomNodeData } from './CustomNode';

interface NodeInspectorDrawerProps {
  node: { id: string; type: string; data: CustomNodeData } | null;
  onClose: () => void;
  onUpdate: (id: string, updatedData: Partial<CustomNodeData>) => void;
  onDelete: (id: string) => void;
}

export const NodeInspectorDrawer: React.FC<NodeInspectorDrawerProps> = ({ node, onClose, onUpdate, onDelete }) => {
  const [label, setLabel] = useState('');
  const [config, setConfig] = useState<Record<string, any>>({});

  useEffect(() => {
    if (node) {
      setLabel(node.data.label || '');
      setConfig(node.data.config || {});
    }
  }, [node]);

  if (!node) return null;

  const nodeType = node.data.type || 'http_request';

  const handleConfigChange = (key: string, value: any) => {
    const newConfig = { ...config, [key]: value };
    setConfig(newConfig);
    onUpdate(node.id, { config: newConfig });
  };

  const handleLabelChange = (newLabel: string) => {
    setLabel(newLabel);
    onUpdate(node.id, { label: newLabel });
  };

  return (
    <div className="fixed right-0 top-16 bottom-0 w-96 bg-slate-900/95 border-l border-white/10 backdrop-blur-xl z-30 shadow-2xl flex flex-col transition-all duration-300">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-white/10 bg-slate-950/50">
        <div className="flex items-center space-x-2">
          <Sliders className="text-indigo-400" size={18} />
          <h3 className="text-base font-semibold text-white">Configure Node</h3>
        </div>
        <div className="flex items-center space-x-2">
          <button
            onClick={() => onDelete(node.id)}
            className="p-1.5 rounded-lg text-rose-400 hover:bg-rose-500/10 hover:text-rose-300 transition-colors"
            title="Delete Node"
          >
            <Trash2 size={18} />
          </button>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-gray-400 hover:bg-white/10 hover:text-white transition-colors"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      {/* Form Content */}
      <div className="flex-1 overflow-y-auto p-5 space-y-6">
        {/* Node ID & Label */}
        <div className="space-y-3">
          <div>
            <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Node ID</label>
            <div className="mt-1 font-mono text-xs text-indigo-300 bg-black/40 px-3 py-2 rounded-lg border border-white/5">
              {node.id}
            </div>
          </div>

          <div>
            <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Node Title / Label</label>
            <input
              type="text"
              value={label}
              onChange={(e) => handleLabelChange(e.target.value)}
              placeholder="Custom Label"
              className="mt-1 w-full bg-slate-950/70 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500 transition-colors"
            />
          </div>
        </div>

        <div className="h-px bg-white/10" />

        {/* Dynamic Config Fields based on Node Type */}
        <div className="space-y-4">
          <h4 className="text-xs font-semibold text-indigo-400 uppercase tracking-wider flex items-center gap-1.5">
            Node Configuration
          </h4>

          {/* HTTP Request Config */}
          {nodeType === 'http_request' && (
            <div className="space-y-4">
              <div>
                <label className="text-xs text-gray-300">HTTP Method</label>
                <select
                  value={config.method || 'GET'}
                  onChange={(e) => handleConfigChange('method', e.target.value)}
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                  <option value="PATCH">PATCH</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-gray-300">Target URL</label>
                <input
                  type="text"
                  value={config.url || ''}
                  onChange={(e) => handleConfigChange('url', e.target.value)}
                  placeholder="https://api.example.com/webhook"
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="text-xs text-gray-300">Headers (JSON format or key-value)</label>
                <textarea
                  value={typeof config.headers === 'string' ? config.headers : JSON.stringify(config.headers || {}, null, 2)}
                  onChange={(e) => {
                    try {
                      const parsed = JSON.parse(e.target.value);
                      handleConfigChange('headers', parsed);
                    } catch {
                      handleConfigChange('headers', e.target.value);
                    }
                  }}
                  rows={3}
                  placeholder='{"Content-Type": "application/json"}'
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg p-3 text-xs text-emerald-300 font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="text-xs text-gray-300">Body Payload</label>
                <textarea
                  value={typeof config.body === 'string' ? config.body : JSON.stringify(config.body || {}, null, 2)}
                  onChange={(e) => handleConfigChange('body', e.target.value)}
                  rows={4}
                  placeholder='{"key": "value"}'
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg p-3 text-xs text-emerald-300 font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>
            </div>
          )}

          {/* Delay Node Config */}
          {nodeType === 'delay' && (
            <div>
              <label className="text-xs text-gray-300">Delay Duration (milliseconds)</label>
              <input
                type="number"
                value={config.duration_ms || config.delay_ms || 1000}
                onChange={(e) => handleConfigChange('duration_ms', parseInt(e.target.value) || 0)}
                placeholder="1000"
                className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-indigo-500"
              />
              <p className="mt-1 text-[11px] text-gray-400">1000 ms = 1 second delay</p>
            </div>
          )}

          {/* Cron Schedule Config */}
          {(nodeType === 'cron' || nodeType === 'schedule') && (
            <div className="space-y-3">
              <div>
                <label className="text-xs text-gray-300">Cron Expression</label>
                <input
                  type="text"
                  value={config.cron || '*/5 * * * *'}
                  onChange={(e) => handleConfigChange('cron', e.target.value)}
                  placeholder="*/5 * * * *"
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-amber-300 font-mono focus:outline-none focus:border-amber-500"
                />
              </div>
              <div className="text-[11px] text-gray-400 space-y-1 bg-black/30 p-2.5 rounded-lg">
                <p>• <code className="text-amber-400">*/5 * * * *</code> : Every 5 minutes</p>
                <p>• <code className="text-amber-400">0 * * * *</code> : Every hour</p>
                <p>• <code className="text-amber-400">0 0 * * *</code> : Midnight daily</p>
              </div>
            </div>
          )}

          {/* Condition Logic Config */}
          {nodeType === 'condition' && (
            <div className="space-y-4">
              <div>
                <label className="text-xs text-gray-300">Target Field / Expression</label>
                <input
                  type="text"
                  value={config.field || ''}
                  onChange={(e) => handleConfigChange('field', e.target.value)}
                  placeholder="e.g. status or output.code"
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="text-xs text-gray-300">Operator</label>
                <select
                  value={config.operator || '=='}
                  onChange={(e) => handleConfigChange('operator', e.target.value)}
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="==">Equals (==)</option>
                  <option value="!=">Not Equals (!=)</option>
                  <option value=">">Greater Than (&gt;)</option>
                  <option value="<">Less Than (&lt;)</option>
                  <option value="contains">Contains</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-gray-300">Comparison Value</label>
                <input
                  type="text"
                  value={config.value || ''}
                  onChange={(e) => handleConfigChange('value', e.target.value)}
                  placeholder="200 or true"
                  className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>
            </div>
          )}

          {/* Logger Config */}
          {nodeType === 'logger' && (
            <div>
              <label className="text-xs text-gray-300">Log Message / Text</label>
              <textarea
                value={config.message || ''}
                onChange={(e) => handleConfigChange('message', e.target.value)}
                rows={3}
                placeholder="Execution logged successfully..."
                className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg p-3 text-xs text-cyan-300 font-mono focus:outline-none focus:border-cyan-500"
              />
            </div>
          )}

          {/* JSON Parser Config */}
          {nodeType === 'json_parser' && (
            <div>
              <label className="text-xs text-gray-300">JSON String / Expression</label>
              <textarea
                value={config.json_string || ''}
                onChange={(e) => handleConfigChange('json_string', e.target.value)}
                rows={4}
                placeholder='{"user_id": 123}'
                className="mt-1 w-full bg-slate-950 border border-white/10 rounded-lg p-3 text-xs text-teal-300 font-mono focus:outline-none focus:border-teal-500"
              />
            </div>
          )}

          {/* Webhook Config */}
          {nodeType === 'webhook' && (
            <div className="space-y-3">
              <div className="p-3 bg-purple-950/30 border border-purple-500/20 rounded-lg text-xs text-purple-300 space-y-1">
                <p className="font-semibold text-purple-200">Webhook Listener Active</p>
                <p className="text-gray-400">
                  Triggers workflow automatically when an HTTP POST payload is sent to:
                </p>
                <p className="font-mono text-purple-300 bg-black/40 p-1.5 rounded border border-purple-500/30 break-all">
                  /webhook/&#123;workflow_id&#125;
                </p>
              </div>
            </div>
          )}

          {/* Manual Config */}
          {nodeType === 'manual' && (
            <div className="p-3 bg-emerald-950/30 border border-emerald-500/20 rounded-lg text-xs text-emerald-300">
              Triggered manually by clicking the <span className="font-semibold text-white">"Run Workflow"</span> button in the canvas bar.
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
