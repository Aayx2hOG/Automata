import React, { useRef, useState } from 'react';
import { X, Trash2, Sliders, Save } from 'lucide-react';
import { CustomNodeData } from './CustomNode';

interface NodeInspectorDrawerProps {
  node: { id: string; type: string; data: CustomNodeData };
  onClose: () => void;
  onUpdate: (id: string, updatedData: Partial<CustomNodeData>) => void;
  onDelete: (id: string) => void;
  onSaveChanges: () => Promise<void>;
  onDiscardChanges: (id: string, data: CustomNodeData) => void;
}

export const NodeInspectorDrawer: React.FC<NodeInspectorDrawerProps> = ({
  node,
  onClose,
  onUpdate,
  onDelete,
  onSaveChanges,
  onDiscardChanges,
}) => {
  const [label, setLabel] = useState(node?.data.label || '');
  const [config, setConfig] = useState<Record<string, any>>(node?.data.config || {});
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const committedData = useRef<CustomNodeData>(node.data);

  const nodeType = node.data.type || 'http_request';

  const handleConfigChange = (key: string, value: any) => {
    const newConfig = { ...config, [key]: value };
    setConfig(newConfig);
    onUpdate(node.id, { config: newConfig });
    setHasUnsavedChanges(true);
  };

  const handleLabelChange = (newLabel: string) => {
    setLabel(newLabel);
    onUpdate(node.id, { label: newLabel });
    setHasUnsavedChanges(true);
  };

  const handleSaveChanges = async () => {
    setIsSaving(true);
    setSaveError(null);
    try {
      await onSaveChanges();
      setHasUnsavedChanges(false);
      committedData.current = { ...committedData.current, label, config };
    } catch (err) {
      setSaveError((err as Error).message || 'Unable to save changes.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleClose = () => {
    if (hasUnsavedChanges && committedData.current) {
      onDiscardChanges(node.id, committedData.current);
    }
    onClose();
  };

  return (
    <div className="fixed right-0 top-16 bottom-0 w-96 bg-[#161a23] border-l border-white/10 backdrop-blur-xl z-30 shadow-2xl flex flex-col transition-all duration-300">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-white/10 bg-[#10141d]">
        <div className="flex items-center space-x-2">
          <Sliders className="text-[#ff6d5a]" size={18} />
          <h3 className="text-sm font-bold text-white uppercase tracking-wider">Node Inspector</h3>
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
            onClick={handleClose}
            className="p-1.5 rounded-lg text-gray-400 hover:bg-white/10 hover:text-white transition-colors"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      {/* Form Content */}
      <div className="flex-1 overflow-y-auto p-5 space-y-6">
        <div className="space-y-3">
          <div>
            <label className="text-[11px] font-bold text-gray-400 uppercase tracking-wider">Node ID</label>
            <div className="mt-1 font-mono text-xs text-[#ff6d5a] bg-[#0d1017] px-3 py-2 rounded-lg border border-white/5">
              {node.id}
            </div>
          </div>

          <div>
            <label className="text-[11px] font-bold text-gray-400 uppercase tracking-wider">Node Title / Label</label>
            <input
              type="text"
              value={label}
              onChange={(e) => handleLabelChange(e.target.value)}
              placeholder="Custom Label"
              className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#ff6d5a] transition-colors"
            />
          </div>
        </div>

        <div className="h-px bg-white/10" />

        {/* Dynamic Config Fields */}
        <div className="space-y-4">
          <h4 className="text-xs font-bold text-[#ff6d5a] uppercase tracking-wider">Parameters</h4>

          {nodeType === 'http_request' && (
            <div className="space-y-4">
              <div>
                <label className="text-xs text-gray-300 font-semibold">HTTP Method</label>
                <select
                  value={config.method || 'GET'}
                  onChange={(e) => handleConfigChange('method', e.target.value)}
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#ff6d5a]"
                >
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                  <option value="PATCH">PATCH</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-gray-300 font-semibold">Target URL</label>
                <input
                  type="text"
                  value={config.url || ''}
                  onChange={(e) => handleConfigChange('url', e.target.value)}
                  placeholder="https://api.example.com/endpoint"
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-[#ff6d5a]"
                />
              </div>

              <div>
                <label className="text-xs text-gray-300 font-semibold">Headers (JSON)</label>
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
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg p-3 text-xs text-cyan-300 font-mono focus:outline-none focus:border-[#ff6d5a]"
                />
              </div>

              <div>
                <label className="text-xs text-gray-300 font-semibold">Body Payload</label>
                <textarea
                  value={typeof config.body === 'string' ? config.body : JSON.stringify(config.body || {}, null, 2)}
                  onChange={(e) => handleConfigChange('body', e.target.value)}
                  rows={4}
                  placeholder='{"key": "value"}'
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg p-3 text-xs text-cyan-300 font-mono focus:outline-none focus:border-[#ff6d5a]"
                />
              </div>
            </div>
          )}

          {nodeType === 'delay' && (
            <div>
              <label className="text-xs text-gray-300 font-semibold">Delay Duration (seconds)</label>
              <input
                type="number"
                value={config.seconds ?? 1}
                onChange={(e) => handleConfigChange('seconds', Number(e.target.value) || 0)}
                placeholder="1000"
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-[#ff6d5a]"
              />
            </div>
          )}

          {(nodeType === 'cron' || nodeType === 'schedule') && (
            <div>
              <label className="text-xs text-gray-300 font-semibold">Cron Expression</label>
              <input
                type="text"
                value={config.cron || '*/5 * * * *'}
                onChange={(e) => handleConfigChange('cron', e.target.value)}
                placeholder="*/5 * * * *"
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-[#ff6d5a] font-mono focus:outline-none focus:border-[#ff6d5a]"
              />
            </div>
          )}

          {nodeType === 'condition' && (
            <div className="space-y-4">
              <div>
                <label className="text-xs text-gray-300 font-semibold">Target Field</label>
                <input
                  type="text"
                  value={config.source_field || ''}
                  onChange={(e) => handleConfigChange('source_field', e.target.value)}
                  placeholder="status"
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-[#ff6d5a]"
                />
              </div>

              <div>
                <label className="text-xs text-gray-300 font-semibold">Operator</label>
                <select
                  value={config.operator || 'equals'}
                  onChange={(e) => handleConfigChange('operator', e.target.value)}
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#ff6d5a]"
                >
                  <option value="equals">Equals</option>
                  <option value="not_equal">Not equal</option>
                  <option value="exists">Exists</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-gray-300 font-semibold">Comparison Value</label>
                <input
                  type="text"
                  value={config.value || ''}
                  onChange={(e) => handleConfigChange('value', e.target.value)}
                  placeholder="200"
                  className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-[#ff6d5a]"
                />
              </div>
            </div>
          )}

          {nodeType === 'logger' && (
            <div>
              <label className="text-xs text-gray-300 font-semibold">Log Message</label>
              <textarea
                value={config.message || ''}
                onChange={(e) => handleConfigChange('message', e.target.value)}
                rows={3}
                placeholder="Workflow execution step completed..."
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-lg p-3 text-xs text-blue-300 font-mono focus:outline-none focus:border-[#ff6d5a]"
              />
            </div>
          )}

          {(nodeType === 'condition' || nodeType === 'json_parser') && (
            <div className="space-y-3">
              <label className="block text-xs text-gray-300">Source node ID
                <input value={config.source_node || ''} onChange={e => handleConfigChange('source_node', e.target.value)} className="mt-1 w-full bg-[#0d1017] rounded p-2" placeholder="http_req_1" />
              </label>
              {nodeType === 'json_parser' && <label className="block text-xs text-gray-300">Source field (JSON array)
                <input value={config.source_field || ''} onChange={e => handleConfigChange('source_field', e.target.value)} className="mt-1 w-full bg-[#0d1017] rounded p-2" placeholder="body" />
              </label>}
            </div>
          )}

          {nodeType === 'webhook_trigger' && (
            <div className="p-3 bg-[#ff6d5a]/10 border border-[#ff6d5a]/30 rounded-lg text-xs text-gray-200 space-y-1">
              <p className="font-bold text-[#ff6d5a]">Signed Webhook Trigger</p>
              <p className="text-gray-400">Triggers on HTTP POST to:</p>
              <p className="font-mono text-[#ff6d5a] bg-[#0d1017] p-1.5 rounded border border-white/10 break-all">
                /webhook/&#123;workflow_id&#125;
              </p>
            </div>
          )}
        </div>
      </div>

      {hasUnsavedChanges && (
        <div className="border-t border-white/10 bg-[#10141d] p-4 space-y-3">
          {saveError && <p role="alert" className="text-xs text-rose-300">{saveError}</p>}
          <p className="text-xs text-amber-300">You have unsaved inspector changes.</p>
          <button
            onClick={handleSaveChanges}
            disabled={isSaving}
            className="w-full inline-flex items-center justify-center space-x-2 px-4 py-2.5 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-sm font-bold shadow-lg shadow-[#ff6d5a]/20 transition-all disabled:cursor-wait disabled:opacity-60"
          >
            <Save size={16} />
            <span>{isSaving ? 'Saving Changes...' : 'Save Changes'}</span>
          </button>
        </div>
      )}
    </div>
  );
};
