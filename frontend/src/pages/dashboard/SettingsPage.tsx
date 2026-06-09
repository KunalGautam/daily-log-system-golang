import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { userApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Loader2, Save, Bell, Palette, Globe, Clock } from 'lucide-react';
import toast from 'react-hot-toast';
import { useEffect } from 'react';

export default function SettingsPage() {
  const queryClient = useQueryClient();

  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: () => userApi.getSettings(),
  });

  const { register, handleSubmit, reset, formState: { isSubmitting } } = useForm();

  useEffect(() => {
    if (settings) {
      reset(settings);
    }
  }, [settings, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: any) => userApi.updateSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      toast.success('Settings saved');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to save'),
  });

  const onSubmit = (data: any) => {
    updateMutation.mutate(data);
  };

  if (isLoading) {
    return (
      <div className="page-container flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="page-container max-w-2xl">
      <div className="page-header">
        <h1 className="page-title">Settings</h1>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Palette className="w-5 h-5 text-primary" />
            <h3 className="text-lg font-semibold text-slate-100">Appearance</h3>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Theme</label>
                <p className="text-xs text-slate-400">Choose your preferred theme</p>
              </div>
              <select {...register('theme')} className="input w-auto">
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Language</label>
                <p className="text-xs text-slate-400">Interface language</p>
              </div>
              <select {...register('language')} className="input w-auto">
                <option value="en">English</option>
              </select>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Bell className="w-5 h-5 text-primary" />
            <h3 className="text-lg font-semibold text-slate-100">Notifications</h3>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Enable Notifications</label>
                <p className="text-xs text-slate-400">Receive reminders and digests</p>
              </div>
              <input type="checkbox" {...register('notification_enabled')} className="w-4 h-4 rounded border-slate-700 bg-slate-800 text-primary focus:ring-primary" />
            </div>
            <div>
              <label className="label">Daily Reminder Time</label>
              <input type="time" {...register('daily_reminder_time')} className="input w-40" />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Weekly Digest</label>
                <p className="text-xs text-slate-400">Receive a weekly summary</p>
              </div>
              <input type="checkbox" {...register('weekly_digest_enabled')} className="w-4 h-4 rounded border-slate-700 bg-slate-800 text-primary focus:ring-primary" />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Monthly Digest</label>
                <p className="text-xs text-slate-400">Receive a monthly summary</p>
              </div>
              <input type="checkbox" {...register('monthly_digest_enabled')} className="w-4 h-4 rounded border-slate-700 bg-slate-800 text-primary focus:ring-primary" />
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Globe className="w-5 h-5 text-primary" />
            <h3 className="text-lg font-semibold text-slate-100">Regional</h3>
          </div>
          <div className="space-y-4">
            <div>
              <label className="label">Timezone</label>
              <select {...register('timezone')} className="input">
                <option value="UTC">UTC</option>
                <option value="America/New_York">Eastern Time</option>
                <option value="America/Chicago">Central Time</option>
                <option value="America/Denver">Mountain Time</option>
                <option value="America/Los_Angeles">Pacific Time</option>
                <option value="Europe/London">London</option>
                <option value="Europe/Paris">Paris</option>
                <option value="Europe/Berlin">Berlin</option>
                <option value="Asia/Tokyo">Tokyo</option>
                <option value="Asia/Shanghai">Shanghai</option>
                <option value="Asia/Kolkata">Kolkata</option>
                <option value="Australia/Sydney">Sydney</option>
              </select>
            </div>
            <div>
              <label className="label">Date Format</label>
              <select {...register('date_format')} className="input">
                <option value="2006-01-02">YYYY-MM-DD</option>
                <option value="01/02/2006">MM/DD/YYYY</option>
                <option value="02-01-2006">DD-MM-YYYY</option>
              </select>
            </div>
            <div>
              <label className="label">Week Start Day</label>
              <select {...register('week_start_day', { valueAsNumber: true })} className="input">
                <option value={0}>Sunday</option>
                <option value={1}>Monday</option>
              </select>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Globe className="w-5 h-5 text-primary" />
            <h3 className="text-lg font-semibold text-slate-100">Privacy</h3>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Public Profile</label>
                <p className="text-xs text-slate-400">Allow others to see your profile</p>
              </div>
              <input type="checkbox" {...register('public_profile')} className="w-4 h-4 rounded" />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <label className="label mb-0">Show on Timeline</label>
                <p className="text-xs text-slate-400">Show your public entries on the timeline</p>
              </div>
              <input type="checkbox" {...register('show_on_timeline')} className="w-4 h-4 rounded" />
            </div>
          </div>
        </div>

        <button type="submit" disabled={isSubmitting} className="btn-primary">
          {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          Save All Settings
        </button>
      </form>
    </div>
  );
}
