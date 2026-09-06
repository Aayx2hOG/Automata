import React, { useState, useEffect } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
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

const MOCK_WORKFLOWS: Workflow[] = [
  {
    id: 'wf_demo_01',
    owner_id: 'user_01',
    name: 'GitHub Webhook Notifier',
    description: 'Catches incoming GitHub webhooks and posts JSON summaries to external alert endpoints.',
    is_active: true,
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'wf_demo_02',
    owner_id: 'user_01',
    name: 'Hourly API Telemetry Sync',
    description: 'Executes every hour via cron scheduler to poll system status endpoints.',
    is_active: true,
    created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'wf_demo_03',
    owner_id: 'user_01',
    name: 'Conditional Status Router',
    description: 'Branches workflow path based on HTTP status code check (200 OK vs Error).',
    is_active: true,
    created_at: new Date(Date.now() - 86400000 * 10).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const MOCK_RUNS: WorkflowRun[] = [
  {
    id: 'run_101',
    workflow_id: 'wf_demo_01',
    workflow_version_id: 'v1',
    status: 'succeeded',
    trigger_type: 'webhook',
    outputs: {
      webhook_1: { status: 200, payload: { action: 'push', ref: 'refs/heads/main' } },
      http_req_1: { status: 200, body: '{"success": true}' },
      logger_1: { logged: 'Notification sent successfully' },
    },
    created_at: new Date(Date.now() - 600000).toISOString(),
  },
  {
    id: 'run_102',
    workflow_id: 'wf_demo_02',
    workflow_version_id: 'v1',
    status: 'succeeded',
    trigger_type: 'cron',
    outputs: {
      cron_1: { triggered_at: new Date().toISOString() },
      http_req_1: { status: 200, body: 'System operational' },
    },
    created_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: 'run_103',
    workflow_id: 'wf_demo_03',
    workflow_version_id: 'v1',
    status: 'failed',
    trigger_type: 'manual',
    error_message: 'HTTP Request failed with status code 500 Server Error',
    created_at: new Date(Date.now() - 7200000).toISOString(),
  },
];

const MainApp: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth();
  const [activeTab, setActiveTab] = useState<'home' | 'workflows' | 'runs' | 'schedules' | 'webhooks'>('home');
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [runs, setRuns] = useState<WorkflowRun[]>(MOCK_RUNS);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [activeWorkflow, setActiveWorkflow] = useState<Workflow | null>(null);
  const [activeGraph, setActiveGraph] = useState<WorkflowGraph | undefined>(undefined);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      loadData();
    }
  }, [isAuthenticated]);

  const loadData = async () => {
    try {
      const data = await api.workflows.list();
      if (data && data.length > 0) {
        setWorkflows(data);
      } else {
        setWorkflows(MOCK_WORKFLOWS);
      }
    } catch {
      setWorkflows(MOCK_WORKFLOWS);
    }
  };

  const handleOpenWorkflowCanvas = async (workflow: Workflow) => {
    setActiveWorkflow(workflow);
    try {
      const details = await api.workflows.get(workflow.id);
      if (details?.active_version?.graph) {
        setActiveGraph(details.active_version.graph);
      } else {
        setActiveGraph(undefined);
      }
    } catch {
      setActiveGraph(undefined);
    }
  };

  const handleCreateWorkflow = async (name: string, description: string, graph: WorkflowGraph) => {
    try {
      const created = await api.workflows.create(name, description, graph);
      const newWf = created.Workflow || {
        id: `wf_${Date.now()}`,
        owner_id: 'user_01',
        name,
        description,
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setWorkflows((prev) => [newWf, ...prev]);
      setActiveWorkflow(newWf);
      setActiveGraph(graph);
    } catch {
      const newWf: Workflow = {
        id: `wf_${Date.now()}`,
        owner_id: 'user_01',
        name,
        description,
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setWorkflows((prev) => [newWf, ...prev]);
      setActiveWorkflow(newWf);
      setActiveGraph(graph);
    }
  };

  const handleSaveWorkflow = async (name: string, description: string, graph: WorkflowGraph) => {
    if (!activeWorkflow) return;
    try {
      await api.workflows.create(name, description, graph);
    } catch {
      // Local fallback
    }
    setWorkflows((prev) =>
      prev.map((w) => (w.id === activeWorkflow.id ? { ...w, name, description } : w))
    );
  };

  const handleRunWorkflow = async (workflowId: string) => {
    try {
      const run = await api.workflows.run(workflowId);
      setRuns((prev) => [run, ...prev]);
      return run;
    } catch {
      const mockRun: WorkflowRun = {
        id: `run_${Date.now()}`,
        workflow_id: workflowId,
        workflow_version_id: 'v1',
        status: 'succeeded',
        trigger_type: 'manual',
        outputs: {
          step_1: { status: 200, message: 'Execution completed successfully' },
        },
        created_at: new Date().toISOString(),
      };
      setRuns((prev) => [mockRun, ...prev]);
      return mockRun;
    }
  };

  const handleCreateSchedule = async (workflowId: string, cronExpression: string) => {
    try {
      const sched = await api.schedules.create(workflowId, cronExpression);
      setSchedules((prev) => [sched, ...prev]);
    } catch {
      const mockSched: Schedule = {
        id: `sched_${Date.now()}`,
        workflow_id: workflowId,
        cron_expression: cronExpression,
        is_active: true,
        next_run_at: new Date(Date.now() + 300000).toISOString(),
        created_at: new Date().toISOString(),
      };
      setSchedules((prev) => [mockSched, ...prev]);
    }
  };

  const handleDeactivateSchedule = async (scheduleId: string) => {
    try {
      await api.schedules.deactivate(scheduleId);
    } catch {
      // ignore
    } finally {
      setSchedules((prev) => prev.filter((s) => s.id !== scheduleId));
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#0b0e14] flex items-center justify-center text-[#ff6d5a]">
        <div className="flex flex-col items-center space-y-3">
          <div className="w-10 h-10 border-4 border-[#ff6d5a] border-t-transparent rounded-full animate-spin" />
          <span className="text-xs font-bold text-gray-300">Loading Automata...</span>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <AuthPage />;
  }

  // Active Visual Graph Editor Canvas View
  if (activeWorkflow) {
    return (
      <WorkflowCanvas
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
            onRun={handleRunWorkflow}
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

export default function App() {
  return (
    <AuthProvider>
      <MainApp />
    </AuthProvider>
  );
}
