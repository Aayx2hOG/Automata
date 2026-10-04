import React, { useState } from 'react';
import { Webhook, Copy, Send, Check, Terminal, Globe } from 'lucide-react';
import { Workflow } from '../../types';
import { api, API_BASE } from '../../lib/api';

interface WebhookTesterProps {
  workflows: Workflow[];
}

export const WebhookTester: React.FC<WebhookTesterProps> = ({ workflows }) => {
  const [selectedId, setSelectedId] = useState('');
  const selectedWorkflow = workflows.find(w => w.id === selectedId) || workflows[0];
  const [payload, setPayload] = useState('{\n  "event": "user.signup",\n  "user": {\n    "id": 42,\n    "name": "Jane Doe"\n  }\n}');
  const [isSending, setIsSending] = useState(false);
  const [responseOutput, setResponseOutput] = useState<any>(null);
  const [copied, setCopied] = useState(false);

  const getWebhookUrl = (workflowId: string) => {
    return new URL(`${API_BASE}/webhook/${workflowId}`, window.location.origin).href;
  };

  const getCurlSnippet = (workflowId: string) => {
    const quote = (value: string) => "'" + value.replaceAll("'", "'\\''") + "'";
    return `read -r -s -p 'Webhook secret: ' secret; echo
payload=${quote(payload)}
timestamp=$(date +%s)
signature=$(printf '%s.%s' "$timestamp" "$payload" | openssl dgst -sha256 -hmac "$secret" | awk '{print $NF}')
curl ${quote(getWebhookUrl(workflowId))} -H 'Content-Type: application/json' -H "X-Webhook-Timestamp: $timestamp" -H "X-Webhook-Signature: $signature" --data-binary "$payload"`;

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
      const parsed = JSON.parse(payload);
      if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') throw new Error('Payload must be a JSON object.');
      const res = await api.webhook.trigger(selectedWorkflow, payload);
      setResponseOutput({ status: 'Success', result: res });
    } catch (err: any) {
      setResponseOutput({ status: 'Failed', error: err.message || 'Webhook trigger error' });
    } finally {
      setIsSending(false);
    }
  };

  return (
    <div className="space-y-8 font-sans">
      <div>
        <h2 className="text-lg font-bold text-white flex items-center space-x-2">
          <Webhook className="text-[#ff6d5a]" size={22} />
          <span>Webhook Endpoint Console</span>
        </h2>
        <p className="text-xs text-gray-400 mt-1">
          Trigger workflows externally via HTTP POST requests or test payloads directly in browser.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Endpoint Selector & Copy Tool */}
        <div className="p-6 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl space-y-6">
          <div>
            <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Select Workflow</label>
            {workflows.length === 0 ? (
              <div className="mt-1 rounded-xl border border-dashed border-white/15 bg-[#0d1017] px-4 py-3 text-sm text-gray-500">
                No workflows created yet. Create a workflow to test its webhook.
              </div>
            ) : (
              <select
                value={selectedWorkflow?.id || ''}
                onChange={(e) => {
                  const wf = workflows.find((w) => w.id === e.target.value);
                  if (wf) setSelectedId(wf.id);
                }}
                className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-[#ff6d5a]"
              >
                {workflows.map((wf) => (
                  <option key={wf.id} value={wf.id}>
                    {wf.name}
                  </option>
                ))}
              </select>
            )}
          </div>

          {selectedWorkflow && (
            <details className="text-xs text-gray-400"><summary>Reveal webhook signing secret</summary><code className="break-all">{selectedWorkflow.webhook_secret}</code></details>
          )}
          {selectedWorkflow && (
            <div className="space-y-4">
              <div>
                <label className="text-xs font-bold text-gray-400 uppercase tracking-wider">Webhook Endpoint URL</label>
                <div className="mt-1 flex items-center space-x-2 bg-[#0d1017] p-3 rounded-xl border border-white/10 font-mono text-xs text-[#ff6d5a] break-all">
                  <Globe size={16} className="text-[#ff6d5a] shrink-0" />
                  <span className="flex-1 font-bold">{getWebhookUrl(selectedWorkflow.id)}</span>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-xs font-bold text-gray-400 uppercase tracking-wider">cURL Command</label>
                  <button
                    onClick={handleCopyCurl}
                    className="flex items-center space-x-1 text-xs text-[#ff6d5a] hover:text-[#ff8575] font-bold"
                  >
                    {copied ? <Check size={14} /> : <Copy size={14} />}
                    <span>{copied ? 'Copied!' : 'Copy cURL'}</span>
                  </button>
                </div>
                <div className="bg-[#0d1017] p-3 rounded-xl border border-white/10 font-mono text-xs text-emerald-400 overflow-x-auto">
                  <pre>{getCurlSnippet(selectedWorkflow.id)}</pre>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Payload Test Console */}
        <div className="p-6 rounded-2xl bg-[#161a23] border border-white/10 shadow-xl space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-white flex items-center space-x-2">
              <Terminal size={16} className="text-[#ff6d5a]" />
              <span>Test Payload Runner</span>
            </h3>

            <button
              onClick={handleSendTrigger}
              disabled={isSending || !selectedWorkflow}
              className="flex items-center space-x-2 px-4 py-2 rounded-xl bg-[#ff6d5a] hover:bg-[#ff8575] text-white text-xs font-bold shadow-lg shadow-[#ff6d5a]/30 transition-all hover:scale-105 disabled:opacity-50"
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
              className="mt-1 w-full bg-[#0d1017] border border-white/10 rounded-xl p-3 text-xs text-emerald-400 font-mono focus:outline-none focus:border-[#ff6d5a]"
            />
          </div>

          {responseOutput && (
            <div className="space-y-1">
              <label className="text-xs font-bold text-gray-400 uppercase tracking-wider">Response</label>
              <div className="p-3 rounded-xl bg-[#0d1017] border border-white/10 font-mono text-xs text-[#ff6d5a] max-h-36 overflow-y-auto font-bold">
                <pre>{JSON.stringify(responseOutput, null, 2)}</pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
