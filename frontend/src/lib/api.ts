import { AuthTokens, Schedule, User, Workflow, WorkflowGraph, WorkflowRun, WorkflowVersion } from '../types';

export const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

class ApiError extends Error {
  status: number;
  data: any;

  constructor(message: string, status: number, data?: any) {
    super(message);
    this.status = status;
    this.data = data;
  }
}

function getStoredToken(): string | null {
  return localStorage.getItem('automata_access_token');
}

function getStoredRefreshToken(): string | null {
  return localStorage.getItem('automata_refresh_token');
}

export function setTokens(tokens: AuthTokens | null) {
  if (tokens) {
    localStorage.setItem('automata_access_token', tokens.access_token);
    localStorage.setItem('automata_refresh_token', tokens.refresh_token);
  } else {
    localStorage.removeItem('automata_access_token');
    localStorage.removeItem('automata_refresh_token');
  }
}

let refreshing: Promise<AuthTokens> | null = null;

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = getStoredToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  const contentType = response.headers.get('content-type');
  const isJson = contentType && contentType.includes('application/json');
  const data = isJson ? await response.json() : await response.text();

  if (!response.ok) {
    // If 401 Unauthorized and we have a refresh token, attempt refresh
    if (response.status === 401 && getStoredRefreshToken() && !endpoint.includes('/auth/') && !endpoint.startsWith('/webhook/')) {
      let refreshed: AuthTokens;
      try {
        // A slower request may receive its 401 after another request already rotated tokens.
        const currentToken = getStoredToken();
        if (currentToken && currentToken !== token) {
          refreshed = {access_token: currentToken, refresh_token: getStoredRefreshToken()!};
        } else {
        if (!refreshing) refreshing = api.auth.refresh().then(tokens => {setTokens(tokens); return tokens;}).finally(() => {refreshing = null;});
        refreshed = await refreshing;
        }
      } catch (err) {
        if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
          setTokens(null);
          window.dispatchEvent(new Event('automata:session-expired'));
        }
        throw err;
      }
      headers.Authorization = `Bearer ${refreshed.access_token}`;
      const retry = await fetch(`${API_BASE}${endpoint}`, {...options, headers});
      const result = retry.status === 204 ? null : await retry.json();
      if (!retry.ok) throw new ApiError(result?.error || 'Request failed', retry.status, result);
      return result as T;
    }
    const errMsg = typeof data === 'object' ? (data?.errors?.map((e: {node_id?: string; message: string}) => `${e.node_id ? e.node_id + ': ' : ''}${e.message}`).join('; ') || data?.error || data?.message || 'API error') : data;
    throw new ApiError(errMsg, response.status, data);
  }

  return data as T;
}

export const api = {
  auth: {
    register: async (email: string, password: string): Promise<AuthTokens & { user: User }> => {
      const data = await request<AuthTokens & { user: User }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
      setTokens(data);
      return data;
    },
    login: async (email: string, password: string): Promise<AuthTokens> => {
      const tokens = await request<AuthTokens>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
      setTokens(tokens);
      return tokens;
    },
    refresh: async (): Promise<AuthTokens> => {
      const refreshToken = getStoredRefreshToken();
      if (!refreshToken) throw new ApiError('No refresh token available', 401);
      return request<AuthTokens>('/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    },
    logout: async (): Promise<void> => {
      const refreshToken = getStoredRefreshToken();
      try {
        if (refreshToken) {
          await request('/auth/logout', {
            method: 'POST',
            body: JSON.stringify({ refresh_token: refreshToken }),
          });
        }
      } finally {
        setTokens(null);
      }
    },
    me: async (): Promise<User> => {
      return request<User>('/users/me');
    },
  },

  workflows: {
    list: async (): Promise<Workflow[]> => {
      const res = await request<{ workflows: Workflow[] }>('/workflows');
      return res.workflows || [];
    },
    get: async (id: string): Promise<{ workflow: Workflow; active_version?: WorkflowVersion }> => {
      return request<{ workflow: Workflow; active_version?: WorkflowVersion }>(`/workflows/${id}`);
    },
    create: async (name: string, description: string, graph: WorkflowGraph): Promise<{ Workflow: Workflow; Version: WorkflowVersion }> => {
      return request<{ Workflow: Workflow; Version: WorkflowVersion }>('/workflows', {
        method: 'POST',
        body: JSON.stringify({ name, description, graph }),
      });
    },
    update: async (id: string, name: string, description: string, graph: WorkflowGraph): Promise<{active_version: WorkflowVersion}> => request(`/workflows/${id}`, {method: 'PUT', body: JSON.stringify({name, description, graph})}),
    run: async (id: string): Promise<WorkflowRun> => {
      return request<WorkflowRun>(`/workflows/${id}/run`, {
        method: 'POST',
      });
    },
  },

  runs: {
    list: async (offset = 0): Promise<WorkflowRun[]> => {
      const data = await request<{runs: WorkflowRun[]}>(`/runs?limit=50&offset=${offset}`);
      return data.runs;
    },
    get: async (id: string): Promise<WorkflowRun> => {
      return request<WorkflowRun>(`/runs/${id}`);
    },
  },

  schedules: {
    list: async (workflowId: string): Promise<Schedule[]> => {
      const res = await request<{ schedules: Schedule[] }>(`/workflows/${workflowId}/schedules`);
      return res.schedules || [];
    },
    create: async (workflowId: string, cronExpression: string): Promise<Schedule> => {
      return request<Schedule>(`/workflows/${workflowId}/schedules`, {
        method: 'POST',
        body: JSON.stringify({ cron_expression: cronExpression }),
      });
    },
    deactivate: async (scheduleId: string): Promise<void> => {
      return request(`/schedules/${scheduleId}`, {
        method: 'DELETE',
      });
    },
  },

  webhook: {
    trigger: async (workflow: Workflow, body: string): Promise<{run_id: string; status: string}> => {
      const timestamp = Math.floor(Date.now() / 1000).toString();
      const encoder = new TextEncoder();
      const key = await crypto.subtle.importKey('raw', encoder.encode(workflow.webhook_secret), {name: 'HMAC', hash: 'SHA-256'}, false, ['sign']);
      const digest = await crypto.subtle.sign('HMAC', key, encoder.encode(`${timestamp}.${body}`));
      const signature = Array.from(new Uint8Array(digest), b => b.toString(16).padStart(2, '0')).join('');
      return request(`/webhook/${workflow.id}`, {method: 'POST', body, headers: {'X-Webhook-Timestamp': timestamp, 'X-Webhook-Signature': signature}});
    },
  },
};
