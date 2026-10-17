import { Home, Layers, PlayCircle, Clock, Webhook } from 'lucide-react';

export const destinations = [
  { id: 'home', label: 'Overview', icon: Home, description: 'Your automation workspace at a glance.' },
  { id: 'workflows', label: 'Workflows', icon: Layers, description: 'Build, organize, and run your automations.' },
  { id: 'runs', label: 'Executions', icon: PlayCircle, description: 'Follow every run and investigate individual steps.' },
  { id: 'schedules', label: 'Schedules', icon: Clock, description: 'Keep recurring work running on time.' },
  { id: 'webhooks', label: 'Webhooks', icon: Webhook, description: 'Connect events to your workflows and test deliveries.' },
] as const;
export type WorkspaceTab = typeof destinations[number]['id'];
