import { useState } from 'react';
import { GraphNode, OutputReference } from '../../types';

const targets: Record<string, string[]> = {
  http_request: ['url', 'body', 'headers'], logger: ['message'], condition: ['value'], delay: ['seconds'],
};
const fields: Record<string, string[]> = {
  http_request: ['status_code', 'body', 'success'], json_parser: ['parsed'], condition: ['result'], logger: ['logged_message'],
};
const inputClass = 'w-full rounded border border-white/10 bg-[#0d1017] p-2 text-xs text-white';

function readPath(output: unknown, path: string[]): unknown {
  let value = output;
  for (const key of path) {
    if (!value || typeof value !== 'object' || !Object.hasOwn(value, key)) return undefined;
    value = (value as Record<string, unknown>)[key];
  }
  return value;
}

export function InputMapping({ type, sources, inputs = {}, outputs, onChange }: {
  type: string;
  sources: GraphNode[];
  inputs?: Record<string, OutputReference>;
  outputs?: Record<string, unknown>;
  onChange: (inputs: Record<string, OutputReference>) => void;
}) {
  const [target, setTarget] = useState(targets[type]?.[0] || '');
  const [source, setSource] = useState(sources[0]?.id || '');
  const [path, setPath] = useState('');
  const sourceNode = sources.find(node => node.id === source);
  const segments = path.replace(/^\//, '').split('/').map(part => part.replace(/~1/g, '/').replace(/~0/g, '~'));
  const preview = outputs && path ? readPath(outputs[source], segments) : undefined;
  const samples = [...new Set([...(fields[sourceNode?.type || ''] || []), ...Object.keys((outputs?.[source] as Record<string, unknown>) || {})])];

  if (!targets[type]) return null;
  return <section className="space-y-2 border-t border-white/10 pt-3 text-xs text-gray-300">
    <h4 className="font-semibold">Input mappings</h4>
    <p className="text-gray-400">Use an upstream output instead of a configured value. Missing fields fail the node before it executes.</p>
    {Object.entries(inputs).map(([field, ref]) => <div key={field} className="flex items-center justify-between gap-2 rounded bg-white/5 p-2">
      <span className="break-all">{field} ← {ref.node_id}/{ref.path.join('/')}</span>
      <button className="text-rose-300" aria-label={`Remove ${field} mapping`} onClick={() => { const next = { ...inputs }; delete next[field]; onChange(next); }}>Remove</button>
    </div>)}
    {sources.length ? <>
      <label className="block">Destination field<select className={inputClass} value={target} onChange={event => setTarget(event.target.value)}>{targets[type].map(field => <option key={field}>{field}</option>)}</select></label>
      <label className="block">Upstream node<select className={inputClass} value={source} onChange={event => { setSource(event.target.value); setPath(''); }}>{sources.map(node => <option key={node.id} value={node.id}>{node.label || node.id}</option>)}</select></label>
      <label className="block">Output field<select className={inputClass} value={samples.includes(path) ? path : ''} onChange={event => setPath(event.target.value)}>
        <option value="">Choose a field or enter a path below</option>{samples.map(field => <option key={field} value={field}>{field}</option>)}
      </select></label>
      <label className="block">Field path<input className={inputClass} value={path} onChange={event => setPath(event.target.value)} placeholder="parsed/customer/id" /></label>
      <p className="text-gray-400">Separate nested fields with /. Array indexes start at 0. Use ~1 for a literal / and ~0 for ~.</p>
      {outputs && path && <pre className="max-h-24 overflow-auto text-gray-400">{preview === undefined ? 'Field unavailable in the latest run.' : JSON.stringify(preview, null, 2)}</pre>}
      <button className="text-cyan-300 disabled:opacity-40" disabled={!sourceNode || !path || !target} onClick={() => onChange({ ...inputs, [target]: { node_id: source, path: segments } })}>Apply mapping</button>
    </> : <p className="text-gray-400">Connect an upstream node to add mappings.</p>}
  </section>;
}
