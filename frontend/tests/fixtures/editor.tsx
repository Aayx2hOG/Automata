import React from 'react';
import ReactDOM from 'react-dom/client';
import { WorkflowCanvas } from '../../src/components/canvas/WorkflowCanvas';
import { WorkflowGraph } from '../../src/types';
import '../../src/index.css';
import '@xyflow/react/dist/style.css';

const graph: WorkflowGraph = {
  nodes: [
    { id: 'source', type: 'http_request', label: 'Fetch customer', config: { url: 'https://api.example.com/customer' }, position: { x: 100, y: 100 } },
    { id: 'target', type: 'http_request', label: 'Send customer', config: { url: 'https://api.example.com/notify' }, position: { x: 100, y: 300 } },
  ],
  edges: [{ from_node_id: 'source', to_node_id: 'target' }],
};

ReactDOM.createRoot(document.getElementById('root')!).render(<WorkflowCanvas
  initialGraph={graph}
  workflowName="Editor regression"
  onBack={() => { document.body.dataset.leftEditor = 'true'; }}
  onSave={async (_name, _description, graph) => { const response = await fetch('/test/save', { method: 'POST', body: JSON.stringify(graph) }); if (!response.ok) throw new Error('Save unavailable'); }}
  onRun={async () => ({ id: 'test-run', status: 'running', nodes: [] })}
/>);
