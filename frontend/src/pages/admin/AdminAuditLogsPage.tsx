import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card } from '@/components/ui/card';
import { FileText, Loader2, Search } from 'lucide-react';
import { format } from 'date-fns';

export default function AdminAuditLogsPage() {
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', page],
    queryFn: async () => {
      const token = localStorage.getItem('access_token');
      const res = await fetch('/api/v1/admin/audit-logs?page=' + page, { headers: { Authorization: `Bearer ${token}` } });
      return res.json();
    },
  });

  return (
    <div>
      <h2 className="text-xl font-semibold text-slate-100 mb-6">Audit Logs</h2>

      {isLoading ? (
        <div className="flex justify-center py-12"><Loader2 className="w-8 h-8 animate-spin text-primary" /></div>
      ) : data?.data ? (
        <div className="space-y-2">
          {data.data.map((log: any) => (
            <div key={log.id} className="card">
              <div className="flex items-center gap-3 mb-1">
                <span className="badge badge-primary text-xs">{log.action}</span>
                <span className="text-xs text-slate-500">{format(new Date(log.created_at), 'MMM d, yyyy HH:mm:ss')}</span>
                <span className="text-xs text-slate-500 ml-auto">{log.ip_address}</span>
              </div>
              {log.details && <p className="text-xs text-slate-400 font-mono mt-1">{log.details}</p>}
            </div>
          ))}
          {data.total_pages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary text-sm">Previous</button>
              <span className="text-sm text-slate-400">Page {page} of {data.total_pages}</span>
              <button onClick={() => setPage(p => Math.min(data.total_pages, p + 1))} disabled={page === data.total_pages} className="btn-secondary text-sm">Next</button>
            </div>
          )}
        </div>
      ) : (
        <div className="card text-center py-12">
          <FileText className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <p className="text-sm text-slate-400">No audit logs yet.</p>
        </div>
      )}
    </div>
  );
}
