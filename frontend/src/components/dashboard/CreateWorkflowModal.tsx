import React, { useState } from 'react';
import { X, Webhook, Clock, GitFork, Plus } from 'lucide-react';
import { WorkflowGraph } from '../../types';

interface CreateWorkflowModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (name: string, description: string, graph: WorkflowGraph) => Promise<void>;
}

const PRESET_TEMPLATES = [
  {
    id: 'webhook_http',
    name: 'Webhook to HTTP Alert',
    description: 'Listens for incoming HTTP Webhook payload and forwards data to an external API endpoint.',
    icon: Webhook,
    color: 'text-[#ff6d5a] border-[#ff6d5a]/40 bg-[#ff6d5a]/10',
    graph: {
      nodes: [
        { id: 'webhook_1', type: 'webhook', config: {}, position: { x: 250, y: 100 } },
        { id: 'http_req_1', type: 'http_request', config: { method: 'POST', url: 'https://httpbin.org/post' }, position: { x: 250, y: 260 } },
        { id: 'logger_1', type: 'logger', config: { message: 'Alert sent successfully' }, position: { x: 250, y: 420 } },
      ],
      edges: [
        { from_node_id: 'webhook_1', to_node_id: 'http_req_1' },
        { from_node_id: 'http_req_1', to_node_id: 'logger_1' },
      ],
    },
  },
  {
    id: 'cron_scheduler',
    name: 'Scheduled Data Fetcher',
    description: 'Runs on a cron schedule (every 5 mins), fetches telemetry data, and logs the execution output.',
    icon: Clock,
    color: 'text-amber-400 border-amber-500/40 bg-amber-950/40',
    graph: {
      nodes: [
        { id: 'cron_1', type: 'cron', config: { cron: '*/5 * * * *' }, position: { x: 250, y: 100 } },
        { id: 'http_req_1', type: 'http_request', config: { method: 'GET', url: 'https://api.github.com/zen' }, position: { x: 250, y: 260 } },
        { id: 'logger_1', type: 'logger', config: { message: 'Scheduled sync completed' }, position: { x: 250, y: 420 } },
      ],
      edges: [
        { from_node_id: 'cron_1', to_node_id: 'http_req_1' },
        { from_node_id: 'http_req_1', to_node_id: 'logger_1' },
      ],
    },
  },
  {
    id: 'conditional_branch',
    name: 'Conditional Status Router',
    description: 'Evaluates HTTP response status code and branches to success or error logger.',
    icon: GitFork,
    color: 'text-[#ff6d5a] border-[#ff6d5a]/40 bg-[#ff6d5a]/10',
    graph: {
      nodes: [
        { id: 'manual_1', type: 'manual', config: {}, position: { x: 300, y: 80 } },
        { id: 'http_req_1', type: 'http_request', config: { method: 'GET', url: 'https://httpbin.org/status/200' }, position: { x: 300, y: 220 } },
        { id: 'condition_1', type: 'condition', config: { field: 'status', operator: '==', value: '200' }, position: { x: 300, y: 380 } },
        { id: 'log_success', type: 'logger', config: { message: 'HTTP OK 200 Received' }, position: { x: 140, y: 540 } },
        { id: 'log_error', type: 'logger', config: { message: 'HTTP Error Branch' }, position: { x: 460, y: 540 } },
      ],
      edges: [
        { from_node_id: 'manual_1', to_node_id: 'http_req_1' },
        { from_node_id: 'http_req_1', to_node_id: 'condition_1' },
        { from_node_id: 'condition_1', to_node_id: 'log_success', condition: 'true' },
        { from_node_id: 'condition_1', to_node_id: 'log_error', condition: 'false' },
      ],
    },
  },
  {
    id: 'blank',
    name: 'Blank Canvas',
    description: 'Start from scratch with a manual trigger node and design your custom DAG workflow.',
    icon: Plus,
    color: 'text-cyan-400 border-cyan-500/40 bg-cyan-950/40',
    graph: {
      nodes: [{ id: 'manual_1', type: 'manual', config: {}, position: { x: 250, y: 150 } }],
      edges: [],
    },
  },
];

export const CreateWorkflowModal: React.FC<CreateWorkflowModalProps> = ({ isOpen, onClose, onCreate }) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState(PRESET_TEMPLATES[0]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsSubmitting(true);
    try {
      await onCreate(name, description, selectedTemplate.graph);
      setName('');
      setDescription('');
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-[#0b0e14]/80 backdrop-blur-md z-50 flex items-center justify-center p-4">
      <div className="bg-[#161a23] border border-white/10 rounded-2xl max-w-2xl w-full p-6 shadow-2xl space-y-6">
        <div className="flex items-center justify-between border-b border-white/10 pb-4">
          <div className="flex items-center space-x-2">
            <div className="p-2 rounded-xl bg-[#ff6d5a]/10 text-[#ff6d5a]">
              <Plus size={20} />
            </div>
            <h2 className="text-base font-bold text-white uppercase tracking-wider">Create New Workflow</h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition-colors">
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-4">
            <div>
              <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Workflow Name *</label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. GitHub Webhook Notifier"
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a]"
              />
            </div>

            <div>
              <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Description</label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional workflow description..."
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a]"
              />
            </div>
          </div>

          <div>
            <label className="text-xs font-bold text-gray-300 uppercase tracking-wider mb-3 block">
              Choose Starter Template
            </label>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {PRESET_TEMPLATES.map((tpl) => {
                const Icon = tpl.icon;
                const isSelected = selectedTemplate.id === tpl.id;
                return (
                  <div
                    key={tpl.id}
                    onClick={() => setSelectedTemplate(tpl)}
                    className={`p-4 rounded-xl border cursor-pointer transition-all ${tpl.color} ${
                      isSelected ? 'ring-2 ring-[#ff6d5a] scale-[1.02]' : 'hover:opacity-90 opacity-70'
                    }`}
                  >
                    <div className="flex items-center space-x-2 font-bold text-sm text-white mb-1">
                      <Icon size={18} />
                      <span>{tpl.name}</span>
                    </div>
                    <p className="text-xs text-gray-300 line-clamp-2">{tpl.description}</p>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="flex items-center justify-end space-x-3 pt-4 border-t border-white/10">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-xl text-xs text-gray-400 hover:text-white transition-colors"
            >
              Cancel
            </button>

            <button
              type="submit"
              disabled={isSubmitting || !name.trim()}
              className="px-5 py-2 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-xs font-bold shadow-lg shadow-[#ff6d5a]/30 transition-all hover:scale-105 disabled:opacity-50"
            >
              {isSubmitting ? 'Creating...' : 'Create & Open Canvas'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
