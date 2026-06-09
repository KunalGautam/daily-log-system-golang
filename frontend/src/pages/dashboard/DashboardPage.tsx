import { useQuery } from '@tanstack/react-query';
import { analyticsApi } from '@/services/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, Target, PenLine, TrendingUp, Loader2 } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { Link } from 'react-router-dom';

export default function DashboardPage() {
  const { data: dashboard, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: analyticsApi.getDashboard,
  });

  if (isLoading) {
    return (
      <div className="page-container flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="page-container">
        <div className="card text-center">
          <p className="text-danger">Failed to load dashboard</p>
        </div>
      </div>
    );
  }

  const stats = [
    { label: 'Today\'s Mood', value: dashboard?.today_mood ? `${dashboard.today_mood}/10` : '--', icon: TrendingUp, color: 'text-yellow-400' },
    { label: 'Total Entries', value: dashboard?.entry_count ?? 0, icon: PenLine, color: 'text-blue-400' },
    { label: 'Active Habits', value: dashboard?.active_habits ?? 0, icon: Activity, color: 'text-green-400' },
    { label: 'Week Entries', value: dashboard?.week_entry_count ?? 0, icon: Target, color: 'text-purple-400' },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Dashboard</h1>
          <p className="page-subtitle">Your life at a glance</p>
        </div>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {stats.map((stat) => (
          <div key={stat.label} className="card">
            <div className="flex items-center justify-between mb-2">
              <stat.icon className={`w-5 h-5 ${stat.color}`} />
            </div>
            <p className="text-2xl font-bold text-slate-100">{stat.value}</p>
            <p className="text-xs text-slate-400 mt-1">{stat.label}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="card">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">Current Streaks</h3>
          {dashboard?.current_streaks && dashboard.current_streaks.length > 0 ? (
            <div className="space-y-3">
              {dashboard.current_streaks.map((streak) => (
                <div key={streak.habit_id} className="flex items-center justify-between">
                  <span className="text-sm text-slate-300">{streak.habit_name}</span>
                  <span className="badge badge-primary">{streak.streak} days</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-slate-500">No active streaks</p>
          )}
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">Upcoming Goals</h3>
          {dashboard?.upcoming_goals && dashboard.upcoming_goals.length > 0 ? (
            <div className="space-y-4">
              {dashboard.upcoming_goals.map((goal) => (
                <div key={goal.id}>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm text-slate-300">{goal.title}</span>
                    <span className="text-xs text-slate-400">{Math.round(goal.progress_percent)}%</span>
                  </div>
                  <div className="w-full h-2 bg-slate-800 rounded-full overflow-hidden">
                    <div className="h-full bg-primary rounded-full transition-all" style={{ width: `${Math.min(goal.progress_percent, 100)}%` }} />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-slate-500">No active goals</p>
          )}
        </div>
      </div>

      <div className="card">
        <h3 className="text-lg font-semibold text-slate-100 mb-4">Recent Entries</h3>
        {dashboard?.recent_entries && dashboard.recent_entries.length > 0 ? (
          <div className="space-y-3">
            {dashboard.recent_entries.map((entry) => (
              <Link key={entry.id} to={`/entries/${entry.id}`} className="flex items-center justify-between p-3 rounded-lg hover:bg-slate-800 transition-colors">
                <div>
                  <p className="text-sm font-medium text-slate-200">{entry.title || 'Untitled'}</p>
                  <p className="text-xs text-slate-400">{entry.entry_type}</p>
                </div>
                {entry.mood_score && (
                  <span className="badge badge-primary">{entry.mood_score}/10</span>
                )}
              </Link>
            ))}
          </div>
        ) : (
          <p className="text-sm text-slate-500">No entries yet. <Link to="/entries" className="text-primary">Create one</Link></p>
        )}
      </div>
    </div>
  );
}
