import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { userApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { User, Loader2, Save } from 'lucide-react';
import { useAuthStore } from '@/store/authStore';
import toast from 'react-hot-toast';
import { useEffect } from 'react';

export default function ProfilePage() {
  const { user, setUser } = useAuthStore();
  const queryClient = useQueryClient();

  const { register, handleSubmit, reset, formState: { isSubmitting } } = useForm();

  useEffect(() => {
    if (user) {
      reset({
        display_name: user.display_name || '',
        username: user.username,
        bio: user.bio || '',
      });
    }
  }, [user, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: any) => userApi.updateProfile(data),
    onSuccess: (result) => {
      setUser(result);
      queryClient.invalidateQueries({ queryKey: ['profile'] });
      toast.success('Profile updated');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to update'),
  });

  const onSubmit = (data: any) => {
    updateMutation.mutate(data);
  };

  return (
    <div className="page-container max-w-2xl">
      <div className="page-header">
        <div>
          <h1 className="page-title">Profile</h1>
          <p className="page-subtitle">Manage your personal information</p>
        </div>
      </div>

      <div className="card mb-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center">
            <User className="w-8 h-8 text-primary" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-slate-100">{user?.display_name || user?.username}</h2>
            <p className="text-sm text-slate-400">{user?.email}</p>
            <span className="badge badge-primary text-xs mt-1">{user?.role}</span>
          </div>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="label">Display Name</label>
              <input {...register('display_name')} className="input" />
            </div>
            <div>
              <label className="label">Username</label>
              <input {...register('username')} className="input" />
            </div>
          </div>
          <div>
            <label className="label">Bio</label>
            <textarea {...register('bio')} className="input min-h-[100px]" placeholder="Tell us about yourself..." />
          </div>
          <button type="submit" disabled={isSubmitting} className="btn-primary">
            {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            Save Changes
          </button>
        </form>
      </div>

      <div className="card">
        <h3 className="text-lg font-semibold text-slate-100 mb-4">Account Info</h3>
        <div className="space-y-3 text-sm">
          <div className="flex justify-between"><span className="text-slate-400">Email</span><span className="text-slate-100">{user?.email}</span></div>
          <div className="flex justify-between"><span className="text-slate-400">Role</span><span className="text-slate-100">{user?.role}</span></div>
          <div className="flex justify-between"><span className="text-slate-400">TOTP 2FA</span><span className={user?.totp_enabled ? 'text-green-400' : 'text-slate-500'}>{user?.totp_enabled ? 'Enabled' : 'Disabled'}</span></div>
          <div className="flex justify-between"><span className="text-slate-400">Passkey</span><span className={user?.passkey_enabled ? 'text-green-400' : 'text-slate-500'}>{user?.passkey_enabled ? 'Enabled' : 'Disabled'}</span></div>
          <div className="flex justify-between"><span className="text-slate-400">Last Login</span><span className="text-slate-100">{user?.last_login_at ? new Date(user.last_login_at).toLocaleString() : 'N/A'}</span></div>
          <div className="flex justify-between"><span className="text-slate-400">Joined</span><span className="text-slate-100">{user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}</span></div>
        </div>
      </div>
    </div>
  );
}
