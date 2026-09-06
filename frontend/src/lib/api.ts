import { AuthTokens, Schedule, User, Workflow, WorkflowGraph, WorkflowRun, WorkflowVersion } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

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
    if (response.status === 401 && getStoredRefreshToken() && !endpoint.includes('/auth/')) {
      try {
        const refreshed = await api.auth.refresh();
        setTokens(refreshed);
        // Retry original request
        headers['Authorization'] = `Bearer ${refreshed.access_token}`;
        const retryRes = await fetch(`${API_BASE}${endpoint}`, { ...options, headers });
        const retryData = retryRes.headers.get('content-type')?.includes('application/json')
          ? await retryRes.json()
          : await retryRes.text();
        if (!retryRes.ok) {
          throw new ApiError(retryData?.error || retryData?.message || 'Request failed', retryRes.status, retryData);
        }
        return retryData as T;
      } catch (err) {
        setTokens(null);
        window.location.reload();
      }
    }
    const errMsg = typeof data === 'object' ? data?.error || data?.message || 'API error' : data;
    throw new ApiError(errMsg, response.status, data);
  }

  return data as T;
}

export const api = {
  auth: {
    register: async (email: string, password: string): Promise<{ user: User; tokens: AuthTokens }> => {
      const data = await request<{ user: User; tokens: AuthTokens }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
      if (data.tokens) setTokens(data.tokens);
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
    run: async (id: string): Promise<WorkflowRun> => {
      return request<WorkflowRun>(`/workflows/${id}/run`, {
        method: 'POST',
      });
    },
  },

  runs: {
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
    trigger: async (workflowId: string, payload: any = {}): Promise<any> => {
      return request(`/webhook/${workflowId}`, {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },
  },
};
