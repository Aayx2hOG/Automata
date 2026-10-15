import ReactDOM from 'react-dom/client';
import { LandingPage } from '../../src/components/home/LandingPage';
import { CreateWorkflowModal } from '../../src/components/dashboard/CreateWorkflowModal';
import { useState } from 'react';
import '../../src/index.css';

export function Preview() {
  const [open, setOpen] = useState(false);
  return <main className="mx-auto max-w-7xl p-8 text-gray-100">
    <LandingPage workflows={[]} runs={[]} onLaunchStudio={() => setOpen(true)} onOpenWorkflow={() => {}} onNavigateTab={tab => { document.body.dataset.tab = tab; }} />
    <CreateWorkflowModal isOpen={open} onClose={() => setOpen(false)} onCreate={async () => {}} />
  </main>;
}
ReactDOM.createRoot(document.getElementById('root')!).render(<Preview />);
