import { useQuery } from '@tanstack/react-query';
import { publicApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Globe, Loader2, PenLine, Users, TrendingUp, BarChart3, ArrowLeft } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { Link } from 'react-router-dom';

export default function PublicStatsPage() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['public-stats'],
    queryFn: () => publicApi.getStats(),
  });

  const moodDistData = stats?.mood_distribution
    ? Object.entries(stats.mood_distribution).map(([score, count]) => ({ score: `${score}/10`, count }))
    : [];

  const entriesByTypeData = stats?.entries_by_type
    ? Object.entries(stats.entries_by_type).map(([type, count]) => ({ type, count }))
    : [];

  return (
    <div className="min-h-screen bg-slate-950">
      <header className="border-b border-slate-800">
        <div className="max-w-4xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <BarChart3 className="w-5 h-5 text-primary" />
            <h1 className="text-lg font-bold text-slate-100">Public Statistics</h1>
          </div>
          <Link to="/" className="text-sm text-slate-400 hover:text-slate-100 flex items-center gap-1">
            <ArrowLeft className="w-4 h-4" /> Home
          </Link>
        </div>
      </header>

      <div className="max-w-4xl mx-auto px-4 py-8">
        {isLoading ? (
          <div className="flex justify-center py-12"><Loader2 className="w-8 h-8 animate-spin text-primary" /></div>
        ) : stats ? (
          <>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
              <div className="card">
                <PenLine className="w-5 h-5 text-blue-400 mb-2" />
                <p className="text-2xl font-bold text-slate-100">{stats.total_entries.toLocaleString()}</p>
                <p className="text-xs text-slate-400">Total Entries</p>
              </div>
              <div className="card">
                <Users className="w-5 h-5 text-green-400 mb-2" />
                <p className="text-2xl font-bold text-slate-100">{stats.total_users.toLocaleString()}</p>
                <p className="text-xs text-slate-400">Active Users</p>
              </div>
              <div className="card">
                <TrendingUp className="w-5 h-5 text-purple-400 mb-2" />
                <p className="text-2xl font-bold text-slate-100">{stats.recent_entries.toLocaleString()}</p>
                <p className="text-xs text-slate-400">Entries (7 days)</p>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div className="card">
                <h3 className="text-lg font-semibold text-slate-100 mb-4">Mood Distribution</h3>
                {moodDistData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={moodDistData}>
                      <XAxis dataKey="score" tickFormatter={(v) => v} stroke="#475569" tick={{ fontSize: 12 }} />
                      <YAxis stroke="#475569" tick={{ fontSize: 12 }} />
                      <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#f1f5f9' }} />
                      <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                ) : <p className="text-sm text-slate-500">No data</p>}
              </div>

              <div className="card">
                <h3 className="text-lg font-semibold text-slate-100 mb-4">Entries by Type</h3>
                {entriesByTypeData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={entriesByTypeData}>
                      <XAxis dataKey="type" stroke="#475569" tick={{ fontSize: 12 }} />
                      <YAxis stroke="#475569" tick={{ fontSize: 12 }} />
                      <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#f1f5f9' }} />
                      <Bar dataKey="count" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                ) : <p className="text-sm text-slate-500">No data</p>}
              </div>
            </div>
          </>
        ) : (
          <div className="card text-center py-12">
            <Globe className="w-12 h-12 text-slate-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-slate-300 mb-2">No statistics available</h3>
            <p className="text-sm text-slate-400">Statistics will appear once entries are created.</p>
          </div>
        )}
      </div>
    </div>
  );
}
