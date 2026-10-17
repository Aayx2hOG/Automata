import { useEffect, useState, useRef } from 'react';
import { UserButton } from '@clerk/react';
import * as Dialog from '@radix-ui/react-dialog';
import { Command } from 'cmdk';
import { Plus, Search, Menu, X, ChevronRight, GitBranch } from 'lucide-react';
import { BrandMark } from './BrandMark';
import { useAuth } from '../../context/useAuth';
import { destinations, type WorkspaceTab } from '../../lib/navigation';
import type { Workflow } from '../../types';

interface HeaderProps {
  activeTab: WorkspaceTab;
  setActiveTab: (tab: WorkspaceTab) => void;
  onNewWorkflow: () => void;
  workflows: Workflow[];
  onOpenWorkflow: (workflow: Workflow) => void;
}

export function Header({ activeTab, setActiveTab, onNewWorkflow, workflows, onOpenWorkflow }: HeaderProps) {
  const { user } = useAuth();
  const searchReturnFocus = useRef<HTMLElement | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  useEffect(() => {
    const handleKey = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault(); setSearchOpen(open => !open);
      }
    };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, []);
  const navigate = (tab: WorkspaceTab) => { setActiveTab(tab); setMenuOpen(false); setSearchOpen(false); };
  const sidebar = <>
    <a className="workspace-brand" aria-label="Automata home" href="#/welcome" onClick={() => setMenuOpen(false)}><span className="brand-mark"><BrandMark size={32} /></span><span>Automata<span className="brand-caption">AUTOMATION PLATFORM</span></span></a>
    <div className="workspace-label">Workspace</div>
    <nav aria-label="Main navigation" className="workspace-nav">
      {destinations.map(({ id, label, icon: Icon }) => <a key={id} href={`#/${id}`} aria-current={activeTab === id ? 'page' : undefined} onClick={event => { event.preventDefault(); navigate(id); }}><Icon size={18} /><span>{label}</span>{id === 'workflows' && <span aria-hidden="true" className="nav-count">{workflows.length}</span>}</a>)}
    </nav>
    <div className="sidebar-bottom"><GitBranch size={19} /><p>From idea to execution<span>Build visually. Run with confidence.</span></p></div>
  </>;
  return <>
    <a href="#workspace-content" className="skip-link" onClick={event => { event.preventDefault(); document.getElementById('workspace-content')?.focus(); }}>Skip to content</a>
    <aside className="workspace-sidebar">{sidebar}</aside>
    <header className="workspace-topbar">
      <Dialog.Root open={menuOpen} onOpenChange={setMenuOpen}><Dialog.Trigger asChild><button className="icon-button mobile-menu" aria-label="Open navigation"><Menu size={20} /></button></Dialog.Trigger><Dialog.Portal><Dialog.Overlay className="dialog-overlay" /><Dialog.Content className="mobile-sidebar" aria-describedby={undefined}><Dialog.Title className="sr-only">Workspace navigation</Dialog.Title><Dialog.Close className="icon-button sidebar-close" aria-label="Close navigation"><X size={18} /></Dialog.Close>{sidebar}</Dialog.Content></Dialog.Portal></Dialog.Root>
      <nav aria-label="Breadcrumb" className="workspace-breadcrumb">
        <a href="#/welcome">Home</a><ChevronRight size={14} aria-hidden="true" />
        {activeTab !== 'home' && <><a href="#/home" onClick={event => { event.preventDefault(); navigate('home'); }}>Overview</a><ChevronRight size={14} aria-hidden="true" /></>}
        <strong aria-current="page">{destinations.find(tab => tab.id === activeTab)?.label}</strong>
      </nav>
      <div className="topbar-actions"><button className="workspace-search" onClick={() => setSearchOpen(true)} aria-label="Search workspace"><Search size={16} /><span>Search workspace…</span><kbd>⌘ / Ctrl K</kbd></button><button className="primary-button" onClick={onNewWorkflow}><Plus size={16} /><span>Add Workflow</span></button>{user && <div className="workspace-account"><span title={user.email}>{user.email}</span><UserButton /></div>}</div>
    </header>
    <Dialog.Root open={searchOpen} onOpenChange={setSearchOpen}><Dialog.Portal><Dialog.Overlay className="dialog-overlay" /><Dialog.Content onOpenAutoFocus={() => { searchReturnFocus.current = document.activeElement as HTMLElement; }} onCloseAutoFocus={event => { event.preventDefault(); searchReturnFocus.current?.focus(); }} className="command-dialog" aria-describedby={undefined}><Dialog.Title className="sr-only">Search workspace</Dialog.Title><Command label="Workspace commands"><div className="command-search"><Search size={19} /><Command.Input placeholder="Search pages and workflows…" /><Dialog.Close className="icon-button" aria-label="Close search"><X size={18} /></Dialog.Close></div><Command.List><Command.Empty>No matching pages or workflows.</Command.Empty><Command.Group heading="Navigate">{destinations.map(({ id, label, icon: Icon }) => <Command.Item key={id} onSelect={() => navigate(id)}><Icon size={17} />{label}<ChevronRight size={14} /></Command.Item>)}</Command.Group><Command.Group heading="Actions"><Command.Item onSelect={() => { setSearchOpen(false); onNewWorkflow(); }}><Plus size={17} />Create new workflow</Command.Item></Command.Group>{workflows.length > 0 && <Command.Group heading="Workflows">{workflows.map(workflow => <Command.Item key={workflow.id} value={`workflow ${workflow.name} ${workflow.id}`} keywords={[workflow.description || '']} onSelect={() => { setSearchOpen(false); onOpenWorkflow(workflow); }}><GitBranch size={17} />{workflow.name}</Command.Item>)}</Command.Group>}</Command.List><div className="command-footer">↑ ↓ to navigate <span>↵ to open</span><span>Esc to close</span></div></Command></Dialog.Content></Dialog.Portal></Dialog.Root>
  </>;
}
