// User types
export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  role: 'admin' | 'user' | 'viewer';
  bio: string;
  avatar_url: string;
  email_verified_at: string | null;
  totp_enabled: boolean;
  passkey_enabled: boolean;
  last_login_at: string;
  created_at: string;
}

export interface UserSettings {
  id: string;
  theme: 'system' | 'light' | 'dark';
  language: string;
  timezone: string;
  week_start_day: number;
  date_format: string;
  time_format: string;
  dark_mode: boolean;
  notification_enabled: boolean;
  daily_reminder_time: string;
  weekly_digest_enabled: boolean;
  monthly_digest_enabled: boolean;
  public_profile: boolean;
  show_on_timeline: boolean;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
  totp_code?: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  display_name?: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface Session {
  id: string;
  user_id: string;
  device_info: string;
  ip_address: string;
  user_agent: string;
  is_revoked: boolean;
  expires_at: string;
  created_at: string;
}

// Entry types
export type EntryType = 'mood' | 'activity' | 'journal' | 'habit_record' | 'goal_progress' | 'sleep' | 'exercise' | 'weight' | 'medication' | 'custom';
export type Visibility = 'private' | 'public' | 'unlisted';

export interface Entry {
  id: string;
  user_id: string;
  entry_type: EntryType;
  title: string;
  description: string;
  markdown_content: string;
  mood_score: number | null;
  activities: string;
  visibility: Visibility;
  is_edited: boolean;
  edit_history: string;
  metadata: string;
  source: string;
  entry_date: string;
  created_at: string;
  updated_at: string;
  tags: Tag[];
  attachments: Attachment[];
  user?: User;
}

export interface EntryFilter {
  entry_type?: string;
  visibility?: string;
  start_date?: string;
  end_date?: string;
  tag_ids?: string[];
  mood_min?: number;
  mood_max?: number;
  search?: string;
  sort?: string;
  sort_dir?: string;
  page?: number;
  page_size?: number;
}

export interface CreateEntryRequest {
  entry_type: EntryType;
  title?: string;
  description?: string;
  markdown_content?: string;
  mood_score?: number;
  activities?: string[];
  visibility?: Visibility;
  tag_ids?: string[];
  entry_date?: string;
  metadata?: Record<string, unknown>;
}

export interface Tag {
  id: string;
  user_id: string;
  name: string;
  color: string;
  icon: string;
  created_at: string;
}

export interface Attachment {
  id: string;
  entry_id: string;
  file_name: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  is_image: boolean;
  width: number | null;
  height: number | null;
  created_at: string;
}

// Habit types
export type HabitType = 'daily' | 'weekly' | 'monthly';

export interface Habit {
  id: string;
  user_id: string;
  name: string;
  description: string;
  habit_type: HabitType;
  frequency: number;
  unit: string;
  target_value: number | null;
  color: string;
  icon: string;
  is_archived: boolean;
  current_streak: number;
  longest_streak: number;
  total_completions: number;
  success_rate: number;
  start_date: string;
  created_at: string;
}

export interface HabitLog {
  id: string;
  habit_id: string;
  user_id: string;
  value: number;
  note: string;
  completed_at: string;
  created_at: string;
}

export interface CreateHabitRequest {
  name: string;
  description?: string;
  habit_type: HabitType;
  frequency?: number;
  unit?: string;
  target_value?: number;
  color?: string;
  icon?: string;
  start_date?: string;
}

// Goal types
export type GoalType = 'daily' | 'weekly' | 'monthly' | 'yearly';

export interface Goal {
  id: string;
  user_id: string;
  title: string;
  description: string;
  goal_type: GoalType;
  target_value: number;
  unit: string;
  current_value: number;
  start_date: string;
  end_date: string | null;
  is_completed: boolean;
  completed_at: string | null;
  color: string;
  icon: string;
  is_archived: boolean;
  progress: GoalProgress[];
  milestones: Milestone[];
  created_at: string;
}

export interface GoalProgress {
  id: string;
  goal_id: string;
  value: number;
  note: string;
  recorded_at: string;
  created_at: string;
}

export interface Milestone {
  id: string;
  goal_id: string;
  title: string;
  description: string;
  target_value: number;
  is_reached: boolean;
  reached_at: string | null;
  created_at: string;
}

// Analytics types
export interface DashboardSummary {
  today_mood: number | null;
  entry_count: number;
  week_entry_count: number;
  current_streaks: { habit_id: string; habit_name: string; streak: number }[];
  recent_entries: { id: string; title: string; entry_type: string; mood_score: number | null; created_at: string }[];
  upcoming_goals: { id: string; title: string; target_value: number; current_value: number; progress_percent: number }[];
  active_habits: number;
}

export interface MoodTrends {
  labels: string[];
  scores: number[];
  average: number;
  stability: number;
  entries: { date: string; score: number; activities: string }[];
}

export interface ActivityCount {
  activity: string;
  count: number;
}

// Notification types
export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  data: string;
  read_at: string | null;
  is_read: boolean;
  created_at: string;
}

// Pagination
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// Public types
export interface PublicStats {
  total_entries: number;
  total_users: number;
  entries_by_type: Record<string, number>;
  mood_distribution: Record<number, number>;
  recent_entries: number;
}

export interface HeatmapData {
  date: string;
  count: number;
}
