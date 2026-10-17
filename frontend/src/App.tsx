import React, { useState, useEffect, useRef } from 'react';
import { AuthProvider } from './context/AuthContext';
import { useAuth } from './context/useAuth';
import { AuthPage } from './components/auth/AuthPage';
import { Header } from './components/common/Header';
import { destinations, type WorkspaceTab } from './lib/navigation';
import { LandingPage } from './components/home/LandingPage';
import { MetricsOverview } from './components/dashboard/MetricsOverview';
import { WorkflowList } from './components/dashboard/WorkflowList';
import { CreateWorkflowModal } from './components/dashboard/CreateWorkflowModal';
import { WorkflowCanvas } from './components/canvas/WorkflowCanvas';
import { RunsList } from './components/runs/RunsList';
import { SchedulesManager } from './components/schedules/SchedulesManager';
import { WebhookTester } from './components/webhooks/WebhookTester';
import { Workflow, WorkflowGraph, WorkflowRun, Schedule, ScheduleInput } from './types';
import { api } from './lib/api';

export const Workspace: React.FC = () => {
  const readTab = (): WorkspaceTab => destinations.find(tab => `#/${tab.id}` === window.location.hash)?.id || 'home';
  const [activeTab, updateActiveTab] = useState<WorkspaceTab>(readTab);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState('');
  const setActiveTab = (tab: WorkspaceTab) => {
    window.location.hash = `/${tab}`;
    updateActiveTab(tab);
  };
  useEffect(() => {
    const onHashChange = () => updateActiveTab(readTab());
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  }, []);
  const deletedWorkflowIds = useRef(new Set<string>());
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
        if (!stopped) {setWorkflows(items.filter(w => !deletedWorkflowIds.current.has(w.id))); setRuns(history.filter(r => !deletedWorkflowIds.current.has(r.workflow_id))); setSchedules(groups.flat().filter(s => s.is_active && !deletedWorkflowIds.current.has(s.workflow_id))); setError(null);}
      } catch (err) { if (!stopped) setError((err as Error).message); }
      finally { if (!stopped) {setLoading(false); timer = setTimeout(refresh, 3000);} }
    };
    void refresh();
    return () => {stopped = true; clearTimeout(timer);};
  }, []);

  const loadData = async () => {
    try {setRuns(await api.runs.list()); setError(null);}
    catch (err) {
      setError((err as Error).message);
      throw err;
    }
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
  const handleDeleteWorkflow = async (workflow: Workflow) => {
    await api.workflows.delete(workflow.id);
    deletedWorkflowIds.current.add(workflow.id);
    setWorkflows(items => items.filter(item => item.id !== workflow.id));
    setRuns(items => items.filter(item => item.workflow_id !== workflow.id));
    setSchedules(items => items.filter(item => item.workflow_id !== workflow.id));
    setSelectedWorkflowId(current => current === workflow.id ? '' : current);
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
  const handleCreateSchedule = async (workflowId: string, cronExpression: ScheduleInput) => {
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
    <div className="workspace-shell min-h-screen text-gray-100 font-sans">
      <Header
        workflows={workflows}
        onOpenWorkflow={handleOpenWorkflowCanvas}
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onNewWorkflow={() => setIsCreateModalOpen(true)}
      />

      <main id="workspace-content" tabIndex={-1} className="workspace-content">
        {activeTab !== 'home' && <div className="page-heading"><p className="eyebrow">AUTOMATION WORKSPACE</p><h1>{destinations.find(tab => tab.id === activeTab)?.label}</h1><p>{destinations.find(tab => tab.id === activeTab)?.description}</p></div>}
        {error && <div role="alert" className="mb-4 p-4 rounded-xl bg-rose-950 text-rose-200">{error}</div>}
        {loading && <p role="status" className="mb-4 text-gray-400">Loading workflows…</p>}
        {activeTab === 'home' && (
          <LandingPage
            workflows={workflows}
            runs={runs}
            loading={loading}
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
            loading={loading}
          />
        )}

        {activeTab === 'workflows' && (
          <WorkflowList
            onDelete={handleDeleteWorkflow}
            workflows={workflows}
            onSelect={handleOpenWorkflowCanvas}
            onRun={id => handleRunWorkflow(id).catch(err => {setError(err.message);})}
            onOpenWebhook={workflow => { setSelectedWorkflowId(workflow.id); setActiveTab('webhooks'); }}
            onOpenSchedule={workflow => { setSelectedWorkflowId(workflow.id); setActiveTab('schedules'); }}
            onNewWorkflow={() => setIsCreateModalOpen(true)}
          />
        )}

        {activeTab === 'runs' && (
          <RunsList runs={runs} onRefresh={loadData} />
        )}

        {activeTab === 'schedules' && (
          <SchedulesManager
            key={selectedWorkflowId}
            initialWorkflowId={selectedWorkflowId}
            workflows={workflows}
            schedules={schedules}
            onCreateSchedule={handleCreateSchedule}
            onDeactivateSchedule={handleDeactivateSchedule}
          />
        )}

        {activeTab === 'webhooks' && (
          <WebhookTester key={selectedWorkflowId} initialWorkflowId={selectedWorkflowId} workflows={workflows} />
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
  const { user, isLoading, error } = useAuth();
  const [showWelcome, setShowWelcome] = useState(() => !destinations.some(tab => window.location.hash === `#/${tab.id}`));
  useEffect(() => {
    const sync = () => setShowWelcome(!destinations.some(tab => window.location.hash === `#/${tab.id}`));
    window.addEventListener('hashchange', sync);
    return () => window.removeEventListener('hashchange', sync);
  }, []);
  // An unresolved session is not a signed-out session. Keep the welcome page
  // unmounted until we know whether the user belongs in the workspace.
  if (isLoading) return null;
  return user && !showWelcome ? <Workspace key={user.id} /> : <AuthPage error={error} onRetry={() => window.location.reload()} />;
};

export default function App() {
  return (
    <AuthProvider>
      <MainApp />
    </AuthProvider>
  );
}
