import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Bell, Loader2, CheckCheck, ExternalLink } from 'lucide-react';
import { format } from 'date-fns';

export default function NotificationsPage() {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => userApi.listNotifications(),
  });

  const markReadMutation = useMutation({
    mutationFn: (id: string) => userApi.markNotificationRead(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });

  const markAllReadMutation = useMutation({
    mutationFn: () => userApi.markAllNotificationsRead(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });

  return (
    <div className="page-container max-w-3xl">
      <div className="page-header">
        <div>
          <h1 className="page-title">Notifications</h1>
          <p className="page-subtitle">Stay updated with your life logging activity</p>
        </div>
        <button onClick={() => markAllReadMutation.mutate()} className="btn-secondary">
          <CheckCheck className="w-4 h-4" /> Mark All Read
        </button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
        </div>
      ) : data?.data && data.data.length > 0 ? (
        <div className="space-y-2">
          {data.data.map((notif) => (
            <div
              key={notif.id}
              className={`card cursor-pointer transition-colors ${!notif.is_read ? 'border-primary/30 bg-slate-900' : ''}`}
              onClick={() => markReadMutation.mutate(notif.id)}
            >
              <div className="flex items-start gap-3">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${!notif.is_read ? 'bg-primary/20' : 'bg-slate-800'}`}>
                  <Bell className={`w-4 h-4 ${!notif.is_read ? 'text-primary' : 'text-slate-500'}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className={`text-sm ${!notif.is_read ? 'font-semibold text-slate-100' : 'text-slate-300'}`}>{notif.title}</h3>
                    <span className="text-xs text-slate-500 shrink-0">
                      {format(new Date(notif.created_at), 'MMM d, HH:mm')}
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 mt-1">{notif.body}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="card text-center py-12">
          <Bell className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-slate-300 mb-2">No notifications</h3>
          <p className="text-sm text-slate-400">You're all caught up!</p>
        </div>
      )}
    </div>
  );
}
