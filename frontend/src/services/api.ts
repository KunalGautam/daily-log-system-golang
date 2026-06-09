import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  Session,
  Entry,
  EntryFilter,
  CreateEntryRequest,
  Tag,
  PaginatedResponse,
  Habit,
  HabitLog,
  CreateHabitRequest,
  Goal,
  GoalProgress,
  Milestone,
  DashboardSummary,
  MoodTrends,
  ActivityCount,
  User,
  UserSettings,
  Notification,
  PublicStats,
  HeatmapData,
} from '../types';

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('access_token');
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    const refreshed = await attemptRefresh();
    if (refreshed) {
      const newToken = localStorage.getItem('access_token');
      (headers as Record<string, string>)['Authorization'] = `Bearer ${newToken}`;
      const retryResponse = await fetch(`${API_BASE}${path}`, {
        ...options,
        headers,
      });
      if (!retryResponse.ok) {
        throw new ApiError(retryResponse.status, await retryResponse.text());
      }
      return retryResponse.json();
    }
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    window.location.href = '/login';
    throw new ApiError(401, 'Unauthorized');
  }

  if (!response.ok) {
    const text = await response.text();
    throw new ApiError(response.status, text);
  }

  if (response.status === 204) return undefined as T;
  return response.json();
}

async function attemptRefresh(): Promise<boolean> {
  const refreshToken = localStorage.getItem('refresh_token');
  if (!refreshToken) return false;

  try {
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) return false;

    const data = await response.json();
    localStorage.setItem('access_token', data.access_token);
    if (data.refresh_token) {
      localStorage.setItem('refresh_token', data.refresh_token);
    }
    return true;
  } catch {
    return false;
  }
}

// Auth API
export const authApi = {
  register: (data: RegisterRequest) => request<AuthResponse>('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data: LoginRequest) => request<AuthResponse>('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  refresh: () => request<AuthResponse>('/auth/refresh', { method: 'POST' }),
  verifyEmail: (token: string) => request<void>(`/auth/verify?token=${token}`, { method: 'POST' }),
  forgotPassword: (email: string) => request<void>('/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email }) }),
  resetPassword: (token: string, password: string) => request<void>('/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) }),
  setupTOTP: () => request<{ secret: string; image: string }>('/auth/totp/setup', { method: 'POST' }),
  verifyTOTP: (code: string) => request<void>('/auth/totp/verify', { method: 'POST', body: JSON.stringify({ code }) }),
  disableTOTP: () => request<void>('/auth/totp/disable', { method: 'POST' }),
  listSessions: () => request<Session[]>('/auth/sessions'),
  revokeSession: (id: string) => request<void>(`/auth/sessions/${id}`, { method: 'DELETE' }),
  revokeAllSessions: () => request<void>('/auth/sessions', { method: 'DELETE' }),
};

// Entries API
export const entriesApi = {
  list: (filter?: EntryFilter) => {
    const params = new URLSearchParams();
    if (filter) {
      Object.entries(filter).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          if (Array.isArray(value)) value.forEach(v => params.append(key, v));
          else params.set(key, String(value));
        }
      });
    }
    return request<PaginatedResponse<Entry>>(`/entries?${params}`);
  },
  get: (id: string) => request<Entry>(`/entries/${id}`),
  create: (data: CreateEntryRequest) => request<Entry>('/entries', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<CreateEntryRequest>) => request<Entry>(`/entries/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/entries/${id}`, { method: 'DELETE' }),
  listTags: () => request<Tag[]>('/entries/tags'),
  createTag: (data: { name: string; color?: string; icon?: string }) => request<Tag>('/entries/tags', { method: 'POST', body: JSON.stringify(data) }),
  deleteTag: (id: string) => request<void>(`/entries/tags/${id}`, { method: 'DELETE' }),
};

// Habits API
export const habitsApi = {
  list: (filter?: { habit_type?: string; is_archived?: boolean; page?: number; page_size?: number }) => {
    const params = new URLSearchParams();
    if (filter) Object.entries(filter).forEach(([k, v]) => { if (v !== undefined) params.set(k, String(v)); });
    return request<PaginatedResponse<Habit>>(`/habits?${params}`);
  },
  get: (id: string) => request<Habit>(`/habits/${id}`),
  create: (data: CreateHabitRequest) => request<Habit>('/habits', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<CreateHabitRequest>) => request<Habit>(`/habits/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/habits/${id}`, { method: 'DELETE' }),
  logCompletion: (id: string, data: { value?: number; note?: string }) => request<HabitLog>(`/habits/${id}/log`, { method: 'POST', body: JSON.stringify(data) }),
  getLogs: (id: string, startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);
    return request<HabitLog[]>(`/habits/${id}/logs?${params}`);
  },
  getStats: () => request<any>('/habits/stats'),
  getHeatmap: (year?: number) => {
    const params = year ? `?year=${year}` : '';
    return request<any[]>(`/habits/heatmap${params}`);
  },
};

// Goals API
export const goalsApi = {
  list: (filter?: { goal_type?: string; is_completed?: boolean; is_archived?: boolean; page?: number; page_size?: number }) => {
    const params = new URLSearchParams();
    if (filter) Object.entries(filter).forEach(([k, v]) => { if (v !== undefined) params.set(k, String(v)); });
    return request<PaginatedResponse<Goal>>(`/goals?${params}`);
  },
  get: (id: string) => request<Goal>(`/goals/${id}`),
  create: (data: any) => request<Goal>('/goals', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: any) => request<Goal>(`/goals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/goals/${id}`, { method: 'DELETE' }),
  addProgress: (id: string, data: { value: number; note?: string }) => request<GoalProgress>(`/goals/${id}/progress`, { method: 'POST', body: JSON.stringify(data) }),
  getProgress: (id: string) => request<GoalProgress[]>(`/goals/${id}/progress`),
  getMilestones: (id: string) => request<Milestone[]>(`/goals/${id}/milestones`),
  createMilestone: (id: string, data: any) => request<Milestone>(`/goals/${id}/milestones`, { method: 'POST', body: JSON.stringify(data) }),
  updateMilestone: (id: string, data: any) => request<Milestone>(`/goals/milestones/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteMilestone: (id: string) => request<void>(`/goals/milestones/${id}`, { method: 'DELETE' }),
  getStats: () => request<any>('/goals/stats'),
};

// Analytics API
export const analyticsApi = {
  getDashboard: () => request<DashboardSummary>('/analytics/dashboard'),
  getMoodTrends: (period?: string, startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (period) params.set('period', period);
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);
    return request<MoodTrends>(`/analytics/mood/trends?${params}`);
  },
  getMoodDistribution: (startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);
    return request<Record<number, number>>(`/analytics/mood/distribution?${params}`);
  },
  getActivityFrequency: (startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);
    return request<ActivityCount[]>(`/analytics/activities/frequency?${params}`);
  },
  getHabitAnalytics: () => request<any>('/analytics/habits'),
  getGoalAnalytics: () => request<any>('/analytics/goals'),
};

// User API
export const userApi = {
  getProfile: () => request<User>('/users/me'),
  updateProfile: (data: Partial<User>) => request<User>('/users/me', { method: 'PUT', body: JSON.stringify(data) }),
  getSettings: () => request<UserSettings>('/users/me/settings'),
  updateSettings: (data: Partial<UserSettings>) => request<UserSettings>('/users/me/settings', { method: 'PUT', body: JSON.stringify(data) }),
  listNotifications: (page?: number) => request<PaginatedResponse<Notification>>(`/users/me/notifications${page ? `?page=${page}` : ''}`),
  markNotificationRead: (id: string) => request<void>(`/users/me/notifications/${id}/read`, { method: 'PUT' }),
  markAllNotificationsRead: () => request<void>('/users/me/notifications/read', { method: 'PUT' }),
};

// Public API
export const publicApi = {
  getFeed: (page?: number) => request<PaginatedResponse<Entry>>(`/public/feed${page ? `?page=${page}` : ''}`),
  getTimeline: (filters?: { entry_type?: string; tag?: string; search?: string; page?: number }) => {
    const params = new URLSearchParams();
    if (filters) Object.entries(filters).forEach(([k, v]) => { if (v !== undefined) params.set(k, String(v)); });
    return request<PaginatedResponse<Entry>>(`/public/timeline?${params}`);
  },
  getStats: () => request<PublicStats>('/public/stats'),
  getHeatmap: (year?: number) => {
    const params = year ? `?year=${year}` : '';
    return request<HeatmapData[]>(`/public/heatmap${params}`);
  },
};

export { ApiError };
