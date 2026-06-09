import { useQuery } from '@tanstack/react-query';
import { Card } from '@/components/ui/card';
import { BarChart3, Loader2, Users, PenLine, Activity, Target } from 'lucide-react';

export default function AdminStatsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      const token = localStorage.getItem('access_token');
      const res = await fetch('/api/v1/admin/stats', { headers: { Authorization: `Bearer ${token}` } });
      return res.json();
    },
  });

  if (isLoading) {
    return <div className="flex justify-center py-12"><Loader2 className="w-8 h-8 animate-spin text-primary" /></div>;
  }

  const stats = [
    { label: 'Total Users', value: data?.total_users ?? 0, icon: Users, color: 'text-blue-400' },
    { label: 'Total Entries', value: data?.total_entries ?? 0, icon: PenLine, color: 'text-green-400' },
    { label: 'Total Habits', value: data?.total_habits ?? 0, icon: Activity, color: 'text-purple-400' },
    { label: 'Total Goals', value: data?.total_goals ?? 0, icon: Target, color: 'text-yellow-400' },
  ];

  return (
    <div>
      <h2 className="text-xl font-semibold text-slate-100 mb-6">System Statistics</h2>
      <div className="grid grid-cols-2 gap-4">
        {stats.map((stat) => (
          <div key={stat.label} className="card">
            <stat.icon className={`w-6 h-6 ${stat.color} mb-2`} />
            <p className="text-2xl font-bold text-slate-100">{stat.value}</p>
            <p className="text-xs text-slate-400">{stat.label}</p>
          </div>
        ))}
      </div>
      {data?.storage_used && (
        <div className="card mt-4">
          <p className="text-sm text-slate-300">Storage Used: <span className="text-slate-100 font-medium">{data.storage_used}</span></p>
        </div>
      )}
    </div>
  );
}
