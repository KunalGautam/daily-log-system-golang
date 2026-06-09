import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card } from '@/components/ui/card';
import { Users, Loader2, Shield, Trash2, Search } from 'lucide-react';
import toast from 'react-hot-toast';

const adminApi = {
  listUsers: async (page = 1) => {
    const token = localStorage.getItem('access_token');
    const res = await fetch('/api/v1/admin/users?page=' + page, { headers: { Authorization: `Bearer ${token}` } });
    return res.json();
  },
  deleteUser: async (id: string) => {
    const token = localStorage.getItem('access_token');
    await fetch('/api/v1/admin/users/' + id, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } });
  },
  updateRole: async (id: string, role: string) => {
    const token = localStorage.getItem('access_token');
    await fetch('/api/v1/admin/users/' + id + '/role', { method: 'PUT', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }, body: JSON.stringify({ role }) });
  },
};

export default function AdminUsersPage() {
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['admin-users', page],
    queryFn: () => adminApi.listUsers(page),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApi.deleteUser(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['admin-users'] }); toast.success('User deleted'); },
    onError: (err: any) => toast.error(err.message || 'Failed to delete'),
  });

  const roleMutation = useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) => adminApi.updateRole(id, role),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['admin-users'] }); toast.success('Role updated'); },
    onError: (err: any) => toast.error(err.message || 'Failed to update role'),
  });

  return (
    <div>
      <h2 className="text-xl font-semibold text-slate-100 mb-6">User Management</h2>

      {isLoading ? (
        <div className="flex justify-center py-12"><Loader2 className="w-8 h-8 animate-spin text-primary" /></div>
      ) : data?.data ? (
        <div className="space-y-3">
          {data.data.map((user: any) => (
            <div key={user.id} className="card flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                  <Users className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium text-slate-100">{user.display_name || user.username}</p>
                  <p className="text-xs text-slate-400">{user.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <select
                  value={user.role}
                  onChange={(e) => roleMutation.mutate({ id: user.id, role: e.target.value })}
                  className="input w-auto text-xs"
                >
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                  <option value="viewer">Viewer</option>
                </select>
                <button onClick={() => deleteMutation.mutate(user.id)} className="p-2 text-slate-500 hover:text-danger transition-colors">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
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
          <Users className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <p className="text-sm text-slate-400">No users found.</p>
        </div>
      )}
    </div>
  );
}
