import { useEffect, useState } from 'react';
import { api } from '../../lib/api';
import { Credential, CredentialInput } from '../../types';

const emptyInput: CredentialInput = { name: '', host: '', header: 'Authorization', value: '' };
const fieldClass = 'mt-1 w-full rounded-lg border border-white/10 bg-[#0d1017] p-2 text-xs text-white';

export function CredentialPicker({ value, onChange }: { value: string; onChange: (id: string) => void }) {
  const [items, setItems] = useState<Credential[]>([]);
  const [input, setInput] = useState<CredentialInput | null>(null);
  const [editingId, setEditingId] = useState<string>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let active = true;
    api.credentials.list().then(items => { if (active) setItems(items); }).catch(error => { if (active) setError(error.message); });
    return () => { active = false; };
  }, []);

  const selected = items.find(item => item.id === value);
  const save = async () => {
    if (!input) return;
    setBusy(true); setError('');
    try {
      const saved = await api.credentials.save(input, editingId);
      setItems(items => [...items.filter(item => item.id !== saved.id), saved]);
      onChange(saved.id); setInput(null);
    } catch (error) { setError((error as Error).message); }
    finally { setBusy(false); }
  };

  const remove = async () => {
    if (!selected || !window.confirm(`Delete “${selected.name}”? Workflows using it will fail until another credential is selected.`)) return;
    setBusy(true); setError('');
    try {
      await api.credentials.delete(selected.id);
      setItems(items => items.filter(item => item.id !== selected.id));
      onChange(''); setInput(null);
    } catch (error) { setError((error as Error).message); }
    finally { setBusy(false); }
  };

  return <section className="space-y-2 border-t border-white/10 pt-3">
    <label className="block text-xs text-gray-300">Stored credential
      <select className={fieldClass} value={value} disabled={busy} onChange={event => { onChange(event.target.value); setInput(null); }}>
        <option value="">None</option>
        {value && !selected && <option value={value}>Unavailable credential</option>}
        {items.map(item => <option key={item.id} value={item.id}>{item.name} · {item.host}</option>)}
      </select>
    </label>
    <div className="flex gap-3 text-xs text-cyan-300">
      <button disabled={busy} onClick={() => { setEditingId(undefined); setInput({ ...emptyInput }); }}>Add credential</button>
      {selected && <>
        <button disabled={busy} onClick={() => { setEditingId(selected.id); setInput({ name: selected.name, host: selected.host, header: selected.header, value: '' }); }}>Replace value</button>
        <button disabled={busy} className="text-rose-300" onClick={remove}>Delete</button>
      </>}
    </div>
    {input && <div className="space-y-2 rounded-lg border border-white/10 p-3">
      <p className="text-xs text-gray-400">Stored separately from the workflow. Sent only to this HTTPS host. Changes apply to every workflow using this credential.</p>
      {(['name', 'host', 'header', 'value'] as const).map(field => <label className="block text-xs text-gray-300" key={field}>
        {{ name: 'Name', host: 'Destination host (for example api.example.com)', header: 'Header name', value: 'Secret value (include Bearer if required)' }[field]}
        <input className={fieldClass} type={field === 'value' ? 'password' : 'text'} autoComplete="off" value={input[field]} disabled={busy} onChange={event => setInput({ ...input, [field]: event.target.value })} />
      </label>)}
      <div className="flex gap-3 text-xs">
        <button className="text-cyan-300" disabled={busy || !input.name || !input.host || !input.header || !input.value} onClick={save}>{busy ? 'Saving…' : 'Save credential'}</button>
        <button className="text-gray-400" disabled={busy} onClick={() => setInput(null)}>Cancel</button>
      </div>
    </div>}
    {error && <p role="alert" className="text-xs text-rose-300">{error}</p>}
  </section>;
}
