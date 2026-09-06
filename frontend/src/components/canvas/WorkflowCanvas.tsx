import React, { useState, useCallback, useMemo } from 'react';
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
  BackgroundVariant,
} from '@xyflow/react';
import { CustomNode, CustomNodeData } from '../nodes/CustomNode';
import { NodeInspectorDrawer } from '../nodes/NodeInspectorDrawer';
import {
  Webhook,
  Clock,
  Play,
  Globe,
  Timer,
  Terminal,
  GitFork,
  Code2,
  Save,
  PlayCircle,
  AlertTriangle,
  Sparkles,
  ArrowLeft,
  CheckCircle2,
  XCircle,
} from 'lucide-react';
import { GraphNode, GraphEdge, WorkflowGraph } from '../../types';

const nodeTypes = {
  custom: CustomNode,
};

interface WorkflowCanvasProps {
  initialGraph?: WorkflowGraph;
  workflowName: string;
  workflowDescription?: string;
  onSave: (name: string, description: string, graph: WorkflowGraph) => Promise<void>;
  onRun: () => Promise<any>;
  onBack: () => void;
}

const PALETTE_ITEMS = [
  { type: 'webhook', title: 'Webhook', category: 'Trigger', icon: Webhook, color: 'text-purple-400 border-purple-500/30 bg-purple-950/30' },
  { type: 'cron', title: 'Cron Schedule', category: 'Trigger', icon: Clock, color: 'text-amber-400 border-amber-500/30 bg-amber-950/30' },
  { type: 'manual', title: 'Manual Trigger', category: 'Trigger', icon: Play, color: 'text-emerald-400 border-emerald-500/30 bg-emerald-950/30' },
  { type: 'http_request', title: 'HTTP Request', category: 'Action', icon: Globe, color: 'text-blue-400 border-blue-500/30 bg-blue-950/30' },
  { type: 'delay', title: 'Delay Pause', category: 'Action', icon: Timer, color: 'text-pink-400 border-pink-500/30 bg-pink-950/30' },
  { type: 'logger', title: 'Logger', category: 'Action', icon: Terminal, color: 'text-cyan-400 border-cyan-500/30 bg-cyan-950/30' },
  { type: 'condition', title: 'Condition Logic', category: 'Logic', icon: GitFork, color: 'text-yellow-400 border-yellow-500/30 bg-yellow-950/30' },
  { type: 'json_parser', title: 'JSON Parser', category: 'Transform', icon: Code2, color: 'text-teal-400 border-teal-500/30 bg-teal-950/30' },
];

export const WorkflowCanvas: React.FC<WorkflowCanvasProps> = ({
  initialGraph,
  workflowName: initialName,
  workflowDescription: initialDesc = '',
  onSave,
  onRun,
  onBack,
}) => {
  const [name, setName] = useState(initialName || 'New Workflow');
  const [description, setDescription] = useState(initialDesc);
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [runResult, setRunResult] = useState<any>(null);
  const [selectedNode, setSelectedNode] = useState<{ id: string; type: string; data: CustomNodeData } | null>(null);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Convert backend WorkflowGraph to React Flow nodes & edges
  const defaultNodes: Node[] = useMemo(() => {
    if (initialGraph?.nodes && initialGraph.nodes.length > 0) {
      return initialGraph.nodes.map((n, idx) => ({
        id: n.id,
        type: 'custom',
        position: n.position || { x: 250 + (idx % 3) * 260, y: 150 + Math.floor(idx / 3) * 160 },
        data: {
          label: n.id,
          type: n.type,
          config: n.config || {},
        },
      }));
    }
    // Default starter canvas with Trigger & Action
    return [
      {
        id: 'webhook_1',
        type: 'custom',
        position: { x: 250, y: 100 },
        data: { label: 'Webhook Trigger', type: 'webhook', config: {} },
      },
      {
        id: 'http_request_1',
        type: 'custom',
        position: { x: 250, y: 280 },
        data: {
          label: 'HTTP Request',
          type: 'http_request',
          config: { method: 'GET', url: 'https://api.github.com/zen' },
        },
      },
    ];
  }, [initialGraph]);

  const defaultEdges: Edge[] = useMemo(() => {
    if (initialGraph?.edges && initialGraph.edges.length > 0) {
      return initialGraph.edges.map((e) => ({
        id: `e-${e.from_node_id}-${e.to_node_id}`,
        source: e.from_node_id,
        target: e.to_node_id,
        sourceHandle: e.condition || 'source',
        label: e.condition ? e.condition.toUpperCase() : undefined,
        style: { stroke: e.condition === 'true' ? '#10b981' : e.condition === 'false' ? '#f43f5e' : '#6366f1' },
      }));
    }
    return [
      {
        id: 'e-webhook_1-http_request_1',
        source: 'webhook_1',
        target: 'http_request_1',
        style: { stroke: '#6366f1' },
      },
    ];
  }, [initialGraph]);

  const [nodes, setNodes, onNodesChange] = useNodesState(defaultNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(defaultEdges);

  const onConnect = useCallback(
    (params: Connection) => {
      const sourceNode = nodes.find((n) => n.id === params.source);
      const isCondition = (sourceNode?.data as any)?.type === 'condition';
      const handleId = params.sourceHandle;

      const newEdge: Edge = {
        ...params,
        id: `e-${params.source}-${params.target}-${handleId || 'source'}`,
        label: isCondition ? (handleId || 'true').toUpperCase() : undefined,
        style: {
          stroke: handleId === 'true' ? '#10b981' : handleId === 'false' ? '#f43f5e' : '#6366f1',
          strokeWidth: 2,
        },
      } as Edge;

      setEdges((eds) => addEdge(newEdge, eds));
    },
    [nodes, setEdges]
  );

  const handleAddNode = (type: string) => {
    const id = `${type}_${Date.now().toString().slice(-4)}`;
    const newNode: Node = {
      id,
      type: 'custom',
      position: {
        x: 300 + Math.random() * 80,
        y: 200 + Math.random() * 80,
      },
      data: {
        label: `${type.toUpperCase()} Node`,
        type,
        config: type === 'http_request' ? { method: 'GET', url: 'https://httpbin.org/get' } : {},
      },
    };
    setNodes((nds) => [...nds, newNode]);
    showToast(`Added ${type} node`);
  };

  const handleNodeClick = (_: React.MouseEvent, node: Node) => {
    setSelectedNode({
      id: node.id,
      type: (node.data as any).type,
      data: node.data as unknown as CustomNodeData,
    });
  };

  const handleUpdateNode = (id: string, updatedData: Partial<CustomNodeData>) => {
    setNodes((nds) =>
      nds.map((n) => {
        if (n.id === id) {
          return {
            ...n,
            data: { ...n.data, ...updatedData },
          };
        }
        return n;
      })
    );
    if (selectedNode && selectedNode.id === id) {
      setSelectedNode((prev) => (prev ? { ...prev, data: { ...prev.data, ...updatedData } } : null));
    }
  };

  const handleDeleteNode = (id: string) => {
    setNodes((nds) => nds.filter((n) => n.id !== id));
    setEdges((eds) => eds.filter((e) => e.source !== id && e.target !== id));
    setSelectedNode(null);
    showToast('Node deleted');
  };

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  // Build graph format for backend API
  const getGraphData = (): WorkflowGraph => {
    const graphNodes: GraphNode[] = nodes.map((n) => ({
      id: n.id,
      type: (n.data as any).type || 'http_request',
      config: (n.data as any).config || {},
      position: { x: n.position.x, y: n.position.y },
    }));

    const graphEdges: GraphEdge[] = edges.map((e) => ({
      from_node_id: e.source,
      to_node_id: e.target,
      condition: e.sourceHandle && e.sourceHandle !== 'source' ? e.sourceHandle : undefined,
    }));

    return { nodes: graphNodes, edges: graphEdges };
  };

  const handleSaveWorkflow = async () => {
    setIsSaving(true);
    try {
      const graph = getGraphData();
      await onSave(name, description, graph);
      showToast('Workflow saved successfully!');
    } catch (err: any) {
      showToast(`Error saving: ${err.message || 'Failed'}`);
    } finally {
      setIsSaving(false);
    }
  };

  const handleRunWorkflow = async () => {
    setIsRunning(true);
    setRunResult(null);
    try {
      // First save graph before executing
      const graph = getGraphData();
      await onSave(name, description, graph);
      const res = await onRun();
      setRunResult(res);
      showToast('Workflow execution triggered!');

      // Highlight status on nodes if outputs available
      if (res && res.outputs) {
        setNodes((nds) =>
          nds.map((n) => ({
            ...n,
            data: {
              ...n.data,
              status: res.status === 'succeeded' ? 'succeeded' : res.status === 'failed' ? 'failed' : 'pending',
              output: res.outputs[n.id],
            },
          }))
        );
      }
    } catch (err: any) {
      showToast(`Run error: ${err.message || 'Failed'}`);
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <div className="relative w-full h-screen bg-slate-950 flex flex-col overflow-hidden text-gray-100">
      {/* Top Navbar Header */}
      <header className="h-16 border-b border-white/10 bg-slate-900/90 backdrop-blur-xl px-5 flex items-center justify-between z-20">
        <div className="flex items-center space-x-4">
          <button
            onClick={onBack}
            className="p-2 rounded-xl text-gray-400 hover:bg-white/10 hover:text-white transition-colors"
            title="Back to Dashboard"
          >
            <ArrowLeft size={20} />
          </button>
          <div>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="bg-transparent font-semibold text-lg text-white focus:outline-none focus:border-b focus:border-indigo-500"
              placeholder="Workflow Name"
            />
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="block bg-transparent text-xs text-gray-400 focus:outline-none"
              placeholder="Add description..."
            />
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center space-x-3">
          <button
            onClick={handleSaveWorkflow}
            disabled={isSaving}
            className="flex items-center space-x-2 px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-sm font-medium text-white border border-white/10 transition-colors disabled:opacity-50"
          >
            <Save size={16} className="text-indigo-400" />
            <span>{isSaving ? 'Saving...' : 'Save Graph'}</span>
          </button>

          <button
            onClick={handleRunWorkflow}
            disabled={isRunning}
            className="flex items-center space-x-2 px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-sm font-semibold text-white shadow-lg shadow-indigo-600/30 transition-all hover:scale-105 active:scale-95 disabled:opacity-50"
          >
            <PlayCircle size={18} />
            <span>{isRunning ? 'Executing...' : 'Run Workflow'}</span>
          </button>
        </div>
      </header>

      {/* Main Content Area */}
      <div className="relative flex-1 flex">
        {/* Left Node Library Palette */}
        <aside className="w-64 border-r border-white/10 bg-slate-900/80 backdrop-blur-md p-4 z-10 flex flex-col space-y-4">
          <div>
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3 flex items-center space-x-1">
              <Sparkles size={14} className="text-indigo-400" />
              <span>Node Library</span>
            </h3>
            <p className="text-xs text-gray-400 mb-4">Click any node below to add it to your DAG workflow graph.</p>
          </div>

          <div className="space-y-2 overflow-y-auto flex-1 pr-1">
            {PALETTE_ITEMS.map((item) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.type}
                  onClick={() => handleAddNode(item.type)}
                  className={`w-full flex items-center space-x-3 p-3 rounded-xl border ${item.color} hover:bg-white/10 transition-all hover:scale-[1.02] text-left group`}
                >
                  <Icon size={18} />
                  <div>
                    <div className="text-xs font-semibold text-white group-hover:text-indigo-300">{item.title}</div>
                    <div className="text-[10px] text-gray-400">{item.category}</div>
                  </div>
                </button>
              );
            })}
          </div>
        </aside>

        {/* Center React Flow Canvas */}
        <div className="flex-1 relative">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={handleNodeClick}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#334155" />
            <Controls />
            <MiniMap maskColor="rgba(15, 23, 42, 0.7)" nodeColor="#6366f1" />
          </ReactFlow>

          {/* Execution Result Banner Overlay */}
          {runResult && (
            <div className="absolute bottom-6 left-6 right-6 max-w-2xl bg-slate-900/95 border border-white/10 p-4 rounded-xl backdrop-blur-xl shadow-2xl z-20 space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  {runResult.status === 'succeeded' ? (
                    <CheckCircle2 className="text-emerald-400" size={20} />
                  ) : (
                    <XCircle className="text-rose-400" size={20} />
                  )}
                  <span className="font-semibold text-sm">
                    Execution {runResult.status?.toUpperCase() || 'COMPLETED'}
                  </span>
                  <span className="text-xs text-gray-400 font-mono">Run ID: {runResult.id}</span>
                </div>
                <button
                  onClick={() => setRunResult(null)}
                  className="text-xs text-gray-400 hover:text-white"
                >
                  Close
                </button>
              </div>

              {runResult.error_message && (
                <div className="text-xs text-rose-300 font-mono bg-rose-950/40 p-2 rounded border border-rose-500/30">
                  Error: {runResult.error_message}
                </div>
              )}

              {runResult.outputs && (
                <div className="text-xs font-mono bg-black/40 p-2 rounded text-emerald-300 max-h-32 overflow-y-auto">
                  <pre>{JSON.stringify(runResult.outputs, null, 2)}</pre>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Right Node Property Inspector Drawer */}
        {selectedNode && (
          <NodeInspectorDrawer
            node={selectedNode}
            onClose={() => setSelectedNode(null)}
            onUpdate={handleUpdateNode}
            onDelete={handleDeleteNode}
          />
        )}
      </div>

      {/* Toast Notification */}
      {toastMessage && (
        <div className="fixed bottom-5 right-5 bg-indigo-600 text-white px-4 py-2.5 rounded-xl shadow-xl text-sm font-medium z-50 animate-bounce">
          {toastMessage}
        </div>
      )}
    </div>
  );
};
