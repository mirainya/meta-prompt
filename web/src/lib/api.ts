const BASE = '/api/v1';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const res = await fetch(`${BASE}${path}`, { ...options, headers });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  return res.json();
}

// Auth
export interface LoginResponse {
  token: string;
  id: number;
  username: string;
  credits: number;
  role: string;
}

export const api = {
  register: (username: string, password: string) =>
    request<{ id: number; username: string; credits: number }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  login: (username: string, password: string) =>
    request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  me: () => request<{ id: number; username: string; credits: number; role: string }>('/user/me'),

  // 获取已启用的 provider 列表
  providers: () => request<ProviderItem[]>('/providers'),

  generate: (input: string, llmProvider?: string) =>
    request<{ id: number }>('/generate', {
      method: 'POST',
      body: JSON.stringify({ input, llm_provider: llmProvider }),
    }),

  histories: (limit = 20, offset = 0) =>
    request<HistoryItem[]>(`/histories?limit=${limit}&offset=${offset}`),

  history: (id: number) => request<HistoryItem>(`/histories/${id}`),

  historyRunning: () => request<RunningStatus>('/histories/running'),

  cancelHistory: (id: number) =>
    request<{ message: string }>(`/histories/${id}/cancel`, { method: 'POST' }),

  // Admin
  adminDashboard: () => request<AdminDashboard>('/admin/dashboard'),

  adminLLMConfigs: () => request<LLMConfigItem[]>('/admin/llm-configs'),

  adminCreateLLMConfig: (data: CreateLLMConfigData) =>
    request<{ message: string }>('/admin/llm-configs', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  adminUpdateLLMConfig: (provider: string, data: Partial<LLMConfigItem>) =>
    request<{ message: string }>(`/admin/llm-configs/${provider}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  adminDeleteLLMConfig: (provider: string) =>
    request<{ message: string }>(`/admin/llm-configs/${provider}`, { method: 'DELETE' }),

  adminTestLLMConfig: (provider: string) =>
    request<{ success: boolean; reply?: string; error?: string }>(`/admin/llm-configs/${provider}/test`, {
      method: 'POST',
    }),

  adminUsers: (limit = 50, offset = 0) =>
    request<{ users: AdminUser[]; total: number }>(`/admin/users?limit=${limit}&offset=${offset}`),

  adminSetCredits: (id: number, credits: number) =>
    request<{ message: string }>(`/admin/users/${id}/credits`, {
      method: 'PUT',
      body: JSON.stringify({ credits }),
    }),

  adminToggleUser: (id: number) =>
    request<{ disabled: boolean }>(`/admin/users/${id}/toggle`, { method: 'PUT' }),

  adminResetPassword: (id: number, password: string) =>
    request<{ message: string }>(`/admin/users/${id}/reset-password`, {
      method: 'PUT',
      body: JSON.stringify({ password }),
    }),

  adminTemplates: (stage?: string) =>
    request<TemplateItem[]>(`/admin/templates${stage ? `?stage=${stage}` : ''}`),

  adminTemplate: (id: number) =>
    request<TemplateItem>(`/admin/templates/${id}`),

  adminUpdateTemplate: (id: number, data: { name: string; description: string; prompt: string }) =>
    request<TemplateItem>(`/admin/templates/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  exportMarkdown: (id: number) =>
    fetch(`${BASE}/histories/${id}/export?format=markdown`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    }).then((r) => r.text()),

  exportJSON: (id: number) =>
    fetch(`${BASE}/histories/${id}/export?format=json`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    }).then((r) => r.text()),

  // API Keys
  listAPIKeys: () => request<APIKeyItem[]>('/api-keys'),
  createAPIKey: (name: string) => request<{ id: number; name: string; key: string; prefix: string }>('/api-keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  }),
  revokeAPIKey: (id: number) => request<{ message: string }>(`/api-keys/${id}`, { method: 'DELETE' }),
  apiKeyStats: (id: number) => request<{ id: number; name: string; prefix: string; rate_limit: number; credits_quota: number; is_active: boolean; last_used_at: string | null; total_calls: number }>(`/api-keys/${id}/stats`),

  // 模板版本
  templateVersions: (id: number) => request<TemplateVersion[]>(`/admin/templates/${id}/versions`),
  templateRollback: (id: number, version: number) => request<TemplateItem>(`/admin/templates/${id}/rollback`, {
    method: 'POST',
    body: JSON.stringify({ version }),
  }),
};

// Types
export interface ProviderItem {
  provider: string;
  model: string;
}

export interface ReviewedPrompt {
  order: number;
  name: string;
  review_passed: boolean;
  prompt_text: string;
  user_instruction: string;
}

export interface GenerateResponse {
  analyzer_output: unknown;
  architect_output: unknown;
  writer_outputs: unknown[];
  reviewer_output: {
    workflow_name: string;
    prompts: ReviewedPrompt[];
  };
  duration_ms: number;
}

export interface HistoryItem {
  id: number;
  user_id: number;
  input: string;
  llm_provider: string;
  status: string;
  current_step: number;
  error_msg?: string;
  reasoner_output: unknown;
  architect_output: unknown;
  generator_output: unknown;
  reviewer_output: unknown;
  template_ids: number[];
  duration_ms: number;
  created_at: string;
  updated_at: string;
}

export interface RunningStatus {
  running: boolean;
  id?: number;
  current_step?: number;
  input?: string;
}

export interface AdminDashboard {
  user_count: number;
  total_generations: number;
  today_generations: number;
}

export interface LLMConfigItem {
  provider: string;
  type: string;
  api_key: string;
  base_url: string;
  model: string;
  max_tokens: number;
  enabled: boolean;
}

export interface CreateLLMConfigData {
  provider: string;
  type: string;
  api_key: string;
  base_url: string;
  model: string;
  max_tokens: number;
  enabled: boolean;
}

export interface AdminUser {
  id: number;
  username: string;
  role: string;
  credits: number;
  disabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface TemplateItem {
  id: number;
  name: string;
  description: string;
  stage: string;
  prompt: string;
  version: number;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface APIKeyItem {
  id: number;
  prefix: string;
  name: string;
  is_active: boolean;
  rate_limit: number;
  last_used_at: string | null;
  created_at: string;
}

export interface TemplateVersion {
  id: number;
  template_id: number;
  prompt: string;
  version: number;
  created_at: string;
}
