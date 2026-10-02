import React, { useState, useEffect } from 'react';
import { AuthProvider } from './context/AuthContext';
import { useAuth } from './context/useAuth';
import { AuthPage } from './components/auth/AuthPage';
import { Header } from './components/common/Header';
import { LandingPage } from './components/home/LandingPage';
import { MetricsOverview } from './components/dashboard/MetricsOverview';
import { WorkflowList } from './components/dashboard/WorkflowList';
import { CreateWorkflowModal } from './components/dashboard/CreateWorkflowModal';
import { WorkflowCanvas } from './components/canvas/WorkflowCanvas';
import { RunsList } from './components/runs/RunsList';
import { SchedulesManager } from './components/schedules/SchedulesManager';
import { WebhookTester } from './components/webhooks/WebhookTester';
import { Workflow, WorkflowGraph, WorkflowRun, Schedule } from './types';
import { api } from './lib/api';

const Workspace: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'home' | 'workflows' | 'runs' | 'schedules' | 'webhooks'>('home');
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [runs, setRuns] = useState<WorkflowRun[]>([]);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [activeWorkflow, setActiveWorkflow] = useState<Workflow | null>(null);
  const [activeGraph, setActiveGraph] = useState<WorkflowGraph | undefined>(undefined);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let stopped = false;
    let timer: ReturnType<typeof setTimeout>;
    const refresh = async () => {
      try {
        const [items, history] = await Promise.all([api.workflows.list(), api.runs.list()]);
        const groups = await Promise.all(items.map(w => api.schedules.list(w.id)));
        if (!stopped) {setWorkflows(items); setRuns(history); setSchedules(groups.flat().filter(s => s.is_active)); setError(null);}
      } catch (err) { if (!stopped) setError((err as Error).message); }
      finally { if (!stopped) {setLoading(false); timer = setTimeout(refresh, 3000);} }
    };
    void refresh();
    return () => {stopped = true; clearTimeout(timer);};
  }, []);

  const loadData = async () => {
    try {setRuns(await api.runs.list()); setError(null);}
    catch (err) {setError((err as Error).message);}
  };
  const handleOpenWorkflowCanvas = async (workflow: Workflow) => {
    setLoading(true); setError(null);
    try {
      const details = await api.workflows.get(workflow.id);
      if (!details.active_version) throw new Error('This workflow has no saved graph.');
      setActiveGraph(details.active_version.graph); setActiveWorkflow(details.workflow);
    } catch (err) {setError((err as Error).message);}
    finally {setLoading(false);}
  };
  const handleCreateWorkflow = async (name: string, description: string, graph: WorkflowGraph) => {
    const created = await api.workflows.create(name, description, graph);
    setWorkflows(prev => [created.Workflow, ...prev]);
    setActiveGraph(created.Version.graph); setActiveWorkflow(created.Workflow);
    setIsCreateModalOpen(false);
  };
  const handleSaveWorkflow = async (name: string, description: string, graph: WorkflowGraph) => {
    if (!activeWorkflow) throw new Error('No workflow selected');
    await api.workflows.update(activeWorkflow.id, name, description, graph);
    setWorkflows(prev => prev.map(w => w.id === activeWorkflow.id ? {...w, name, description} : w));
  };
  const handleRunWorkflow = async (workflowId: string) => {
    const run = await api.workflows.run(workflowId);
    setRuns(prev => [run, ...prev.filter(r => r.id !== run.id)]);
    return run;
  };
  const handleCreateSchedule = async (workflowId: string, cronExpression: string) => {
    const schedule = await api.schedules.create(workflowId, cronExpression);
    setSchedules(prev => [schedule, ...prev]);
  };
  const handleDeactivateSchedule = async (scheduleId: string) => {
    await api.schedules.deactivate(scheduleId);
    setSchedules(prev => prev.filter(s => s.id !== scheduleId));
  };

  // Active Visual Graph Editor Canvas View
  if (activeWorkflow) {
    return (
      <WorkflowCanvas
        key={activeWorkflow.id}
        initialGraph={activeGraph}
        workflowName={activeWorkflow.name}
        workflowDescription={activeWorkflow.description}
        onSave={handleSaveWorkflow}
        onRun={() => handleRunWorkflow(activeWorkflow.id)}
        onBack={() => {
          setActiveWorkflow(null);
          setActiveGraph(undefined);
        }}
      />
    );
  }

  return (
    <div className="min-h-screen bg-[#090d16] text-gray-100 flex flex-col font-sans">
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onNewWorkflow={() => setIsCreateModalOpen(true)}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto p-6 md:p-8">
        {error && <div role="alert" className="mb-4 p-4 rounded-xl bg-rose-950 text-rose-200">{error}</div>}
        {loading && <p role="status" className="mb-4 text-gray-400">Loading workflows…</p>}
        {activeTab === 'home' && (
          <LandingPage
            workflows={workflows}
            onLaunchStudio={() => setIsCreateModalOpen(true)}
            onOpenWorkflow={handleOpenWorkflowCanvas}
            onNavigateTab={setActiveTab}
          />
        )}

        {activeTab !== 'home' && (
          <MetricsOverview
            workflows={workflows}
            runCount={runs.length}
            scheduleCount={schedules.length}
          />
        )}

        {activeTab === 'workflows' && (
          <WorkflowList
            workflows={workflows}
            onSelect={handleOpenWorkflowCanvas}
            onRun={id => handleRunWorkflow(id).catch(err => {setError(err.message);})}
            onOpenWebhook={() => setActiveTab('webhooks')}
            onOpenSchedule={() => setActiveTab('schedules')}
            onNewWorkflow={() => setIsCreateModalOpen(true)}
          />
        )}

        {activeTab === 'runs' && (
          <RunsList runs={runs} onRefresh={loadData} />
        )}

        {activeTab === 'schedules' && (
          <SchedulesManager
            workflows={workflows}
            schedules={schedules}
            onCreateSchedule={handleCreateSchedule}
            onDeactivateSchedule={handleDeactivateSchedule}
          />
        )}

        {activeTab === 'webhooks' && (
          <WebhookTester workflows={workflows} />
        )}
      </main>

      <CreateWorkflowModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onCreate={handleCreateWorkflow}
      />
    </div>
  );
};

const MainApp = () => {
  const { user, isLoading } = useAuth();
  if (isLoading) return <div role="status" className="min-h-screen bg-[#0b0e14] p-8 text-gray-200">Loading Automata…</div>;
  return user ? <Workspace key={user.id} /> : <AuthPage />;
};

export default function App() {
  return (
    <AuthProvider>
      <MainApp />
    </AuthProvider>
  );
}
