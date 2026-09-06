import React, { useState } from 'react';
import { Webhook, Copy, Send, Check, Terminal, Globe } from 'lucide-react';
import { Workflow } from '../../types';
import { api } from '../../lib/api';

interface WebhookTesterProps {
  workflows: Workflow[];
}

export const WebhookTester: React.FC<WebhookTesterProps> = ({ workflows }) => {
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(workflows[0] || null);
  const [payload, setPayload] = useState('{\n  "event": "user.signup",\n  "user": {\n    "id": 42,\n    "name": "Jane Doe"\n  }\n}');
  const [isSending, setIsSending] = useState(false);
  const [responseOutput, setResponseOutput] = useState<any>(null);
  const [copied, setCopied] = useState(false);

  const getWebhookUrl = (workflowId: string) => {
    return `${window.location.origin}/api/webhook/${workflowId}`;
  };

  const getCurlSnippet = (workflowId: string) => {
    return `curl -X POST "${window.location.origin}/api/webhook/${workflowId}" \\
  -H "Content-Type: application/json" \\
  -d '${payload.replace(/\n/g, '')}'`;
  };

  const handleCopyCurl = () => {
    if (!selectedWorkflow) return;
    navigator.clipboard.writeText(getCurlSnippet(selectedWorkflow.id));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSendTrigger = async () => {
    if (!selectedWorkflow) return;
    setIsSending(true);
    setResponseOutput(null);
    try {
      let parsedPayload = {};
      try {
        parsedPayload = JSON.parse(payload);
      } catch {
        parsedPayload = { text: payload };
      }

      const res = await api.webhook.trigger(selectedWorkflow.id, parsedPayload);
      setResponseOutput({ status: 'Success', result: res });
    } catch (err: any) {
      setResponseOutput({ status: 'Failed', error: err.message || 'Webhook trigger error' });
    } finally {
      setIsSending(false);
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-lg font-bold text-white flex items-center space-x-2">
          <Webhook className="text-purple-400" size={22} />
          <span>Webhook Endpoint Inspector & Test Console</span>
        </h2>
        <p className="text-xs text-gray-400 mt-1">
          Trigger workflows externally via HTTP POST requests or test payloads directly in browser.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Endpoint Selector & Copy Tool */}
        <div className="p-6 rounded-2xl bg-slate-900/80 border border-white/10 backdrop-blur-xl shadow-xl space-y-6">
          <div>
            <label className="text-xs font-semibold text-gray-300 uppercase tracking-wider">Select Workflow</label>
            <select
              value={selectedWorkflow?.id || ''}
              onChange={(e) => {
                const wf = workflows.find((w) => w.id === e.target.value);
                if (wf) setSelectedWorkflow(wf);
              }}
              className="mt-1 w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-purple-500"
            >
              {workflows.map((wf) => (
                <option key={wf.id} value={wf.id}>
                  {wf.name}
                </option>
              ))}
            </select>
          </div>

          {selectedWorkflow && (
            <div className="space-y-4">
              <div>
                <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Webhook Endpoint URL</label>
                <div className="mt-1 flex items-center space-x-2 bg-slate-950 p-3 rounded-xl border border-white/10 font-mono text-xs text-purple-300 break-all">
                  <Globe size={16} className="text-purple-400 shrink-0" />
                  <span className="flex-1">{getWebhookUrl(selectedWorkflow.id)}</span>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider">cURL Command</label>
                  <button
                    onClick={handleCopyCurl}
                    className="flex items-center space-x-1 text-xs text-purple-400 hover:text-purple-300"
                  >
                    {copied ? <Check size={14} /> : <Copy size={14} />}
                    <span>{copied ? 'Copied!' : 'Copy cURL'}</span>
                  </button>
                </div>
                <div className="bg-black/60 p-3 rounded-xl border border-white/10 font-mono text-xs text-emerald-400 overflow-x-auto">
                  <pre>{getCurlSnippet(selectedWorkflow.id)}</pre>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Payload Test Console */}
        <div className="p-6 rounded-2xl bg-slate-900/80 border border-white/10 backdrop-blur-xl shadow-xl space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-white flex items-center space-x-2">
              <Terminal size={16} className="text-purple-400" />
              <span>Test Payload Runner</span>
            </h3>

            <button
              onClick={handleSendTrigger}
              disabled={isSending || !selectedWorkflow}
              className="flex items-center space-x-2 px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-xs font-semibold shadow-lg shadow-purple-600/30 transition-all hover:scale-105 disabled:opacity-50"
            >
              <Send size={14} />
              <span>{isSending ? 'Sending...' : 'Send Payload'}</span>
            </button>
          </div>

          <div>
            <label className="text-xs text-gray-400">JSON Payload</label>
            <textarea
              value={payload}
              onChange={(e) => setPayload(e.target.value)}
              rows={6}
              className="mt-1 w-full bg-slate-950 border border-white/10 rounded-xl p-3 text-xs text-emerald-300 font-mono focus:outline-none focus:border-purple-500"
            />
          </div>

          {responseOutput && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Response</label>
              <div className="p-3 rounded-xl bg-black/60 border border-white/10 font-mono text-xs text-purple-300 max-h-36 overflow-y-auto">
                <pre>{JSON.stringify(responseOutput, null, 2)}</pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
