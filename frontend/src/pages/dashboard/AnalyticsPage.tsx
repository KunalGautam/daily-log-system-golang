import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { analyticsApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, LineChart, Line, AreaChart, Area } from 'recharts';
import { TrendingUp, Loader2 } from 'lucide-react';

export default function AnalyticsPage() {
  const [period, setPeriod] = useState('weekly');

  const { data: moodTrends, isLoading: moodLoading } = useQuery({
    queryKey: ['mood-trends', period],
    queryFn: () => analyticsApi.getMoodTrends(period),
  });

  const { data: moodDist } = useQuery({
    queryKey: ['mood-distribution'],
    queryFn: () => analyticsApi.getMoodDistribution(),
  });

  const { data: activities } = useQuery({
    queryKey: ['activity-frequency'],
    queryFn: () => analyticsApi.getActivityFrequency(),
  });

  const { data: habitAnalytics } = useQuery({
    queryKey: ['habit-analytics'],
    queryFn: () => analyticsApi.getHabitAnalytics(),
  });

  const { data: goalAnalytics } = useQuery({
    queryKey: ['goal-analytics'],
    queryFn: () => analyticsApi.getGoalAnalytics(),
  });

  const moodDistData = moodDist ? Object.entries(moodDist).map(([score, count]) => ({ score: `${score}/10`, count })) : [];

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Analytics</h1>
          <p className="page-subtitle">Visualize your life data</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-slate-100">Mood Trends</h3>
            <select value={period} onChange={(e) => setPeriod(e.target.value)} className="input w-auto text-sm">
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
            </select>
          </div>
          {moodLoading ? (
            <div className="flex justify-center py-8"><Loader2 className="w-6 h-6 animate-spin text-primary" /></div>
          ) : moodTrends?.scores && moodTrends.scores.length > 0 ? (
            <>
              <ResponsiveContainer width="100%" height={250}>
                <AreaChart data={moodTrends.labels.map((label, i) => ({ label, score: moodTrends.scores[i] }))}>
                  <defs>
                    <linearGradient id="moodGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <XAxis dataKey="label" stroke="#475569" tick={{ fontSize: 12 }} />
                  <YAxis domain={[0, 10]} stroke="#475569" tick={{ fontSize: 12 }} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#f1f5f9' }}
                  />
                  <Area type="monotone" dataKey="score" stroke="#3b82f6" fill="url(#moodGradient)" strokeWidth={2} />
                </AreaChart>
              </ResponsiveContainer>
              <div className="flex gap-4 mt-3 text-sm">
                <span className="text-slate-400">Average: <strong className="text-slate-100">{moodTrends.average?.toFixed(1)}</strong></span>
                <span className="text-slate-400">Stability: <strong className="text-slate-100">{moodTrends.stability?.toFixed(1)}%</strong></span>
              </div>
            </>
          ) : (
            <p className="text-sm text-slate-500 py-8 text-center">No mood data yet</p>
          )}
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">Mood Distribution</h3>
          {moodDistData.length > 0 ? (
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={moodDistData}>
                <XAxis dataKey="score" stroke="#475569" tick={{ fontSize: 12 }} />
                <YAxis stroke="#475569" tick={{ fontSize: 12 }} />
                <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#f1f5f9' }} />
                <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-slate-500 py-8 text-center">No mood data yet</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="card">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">Top Activities</h3>
          {activities && activities.length > 0 ? (
            <div className="space-y-3">
              {activities.slice(0, 10).map((act) => (
                <div key={act.activity} className="flex items-center justify-between">
                  <span className="text-sm text-slate-300">{act.activity}</span>
                  <span className="text-sm text-slate-400">{act.count}x</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-slate-500">No activity data</p>
          )}
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">Habit Analytics</h3>
          {habitAnalytics ? (
            <div className="space-y-4">
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Total</span>
                <span className="text-slate-100">{habitAnalytics.total_habits}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Active</span>
                <span className="text-green-400">{habitAnalytics.active_habits}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Avg Streak</span>
                <span className="text-slate-100">{habitAnalytics.avg_streak?.toFixed(1) || 0}d</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Avg Completion</span>
                <span className="text-slate-100">{habitAnalytics.avg_completion_rate?.toFixed(0) || 0}%</span>
              </div>
            </div>
          ) : (
            <p className="text-sm text-slate-500">No habit data</p>
          )}
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">Goal Analytics</h3>
          {goalAnalytics ? (
            <div className="space-y-4">
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Total</span>
                <span className="text-slate-100">{goalAnalytics.total_goals}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Completed</span>
                <span className="text-green-400">{goalAnalytics.completed_goals}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">In Progress</span>
                <span className="text-blue-400">{goalAnalytics.in_progress}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Avg Progress</span>
                <span className="text-slate-100">{goalAnalytics.avg_progress?.toFixed(0) || 0}%</span>
              </div>
            </div>
          ) : (
            <p className="text-sm text-slate-500">No goal data</p>
          )}
        </div>
      </div>
    </div>
  );
}
