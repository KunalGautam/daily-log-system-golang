import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { habitsApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Activity, Plus, Check, Loader2, Trash2, Flame } from 'lucide-react';
import { useForm } from 'react-hook-form';
import toast from 'react-hot-toast';

export default function HabitsPage() {
  const [showForm, setShowForm] = useState(false);
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['habits'],
    queryFn: () => habitsApi.list({ page_size: 50 }),
  });

  const { data: stats } = useQuery({
    queryKey: ['habits-stats'],
    queryFn: () => habitsApi.getStats(),
  });

  const { register, handleSubmit, reset, formState: { isSubmitting } } = useForm();

  const createMutation = useMutation({
    mutationFn: (data: any) => habitsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
      queryClient.invalidateQueries({ queryKey: ['habits-stats'] });
      toast.success('Habit created');
      setShowForm(false);
      reset();
    },
    onError: (err: any) => toast.error(err.message || 'Failed to create'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => habitsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
      queryClient.invalidateQueries({ queryKey: ['habits-stats'] });
      toast.success('Habit deleted');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to delete'),
  });

  const logMutation = useMutation({
    mutationFn: ({ id, value }: { id: string; value?: number }) => habitsApi.logCompletion(id, { value }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
      queryClient.invalidateQueries({ queryKey: ['habits-stats'] });
      toast.success('Logged!');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to log'),
  });

  const onSubmit = (formData: any) => {
    createMutation.mutate({
      name: formData.name,
      description: formData.description,
      habit_type: formData.habit_type || 'daily',
      frequency: parseInt(formData.frequency) || 1,
    });
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Habits</h1>
          <p className="page-subtitle">Build and maintain your daily routines</p>
        </div>
        <button onClick={() => setShowForm(!showForm)} className="btn-primary">
          <Plus className="w-4 h-4" /> New Habit
        </button>
      </div>

      {stats && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="card">
            <p className="text-2xl font-bold text-slate-100">{stats.total_habits ?? 0}</p>
            <p className="text-xs text-slate-400">Total Habits</p>
          </div>
          <div className="card">
            <p className="text-2xl font-bold text-green-400">{stats.active_habits ?? 0}</p>
            <p className="text-xs text-slate-400">Active</p>
          </div>
          <div className="card">
            <p className="text-2xl font-bold text-yellow-400">{stats.total_completions ?? 0}</p>
            <p className="text-xs text-slate-400">Completions</p>
          </div>
        </div>
      )}

      {showForm && (
        <div className="card mb-6">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">New Habit</h3>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="flex gap-4">
              <div className="flex-1">
                <label className="label">Name</label>
                <input {...register('name', { required: true })} className="input" placeholder="Read 30 minutes" />
              </div>
              <div className="w-40">
                <label className="label">Type</label>
                <select {...register('habit_type')} className="input">
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>
              <div className="w-32">
                <label className="label">Frequency</label>
                <input type="number" {...register('frequency')} className="input" defaultValue={1} min={1} />
              </div>
            </div>
            <div>
              <label className="label">Description</label>
              <input {...register('description')} className="input" placeholder="Optional description" />
            </div>
            <div className="flex gap-3">
              <button type="submit" disabled={isSubmitting} className="btn-primary">
                {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plus className="w-4 h-4" />}
                Create
              </button>
              <button type="button" onClick={() => setShowForm(false)} className="btn-secondary">Cancel</button>
            </div>
          </form>
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
        </div>
      ) : data?.data && data.data.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {data.data.map((habit) => (
            <div key={habit.id} className="card">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                    <Activity className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-slate-100">{habit.name}</h3>
                    <p className="text-xs text-slate-400">{habit.habit_type} &middot; {habit.frequency}x</p>
                  </div>
                </div>
                <button onClick={() => deleteMutation.mutate(habit.id)} className="text-slate-500 hover:text-danger transition-colors">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>

              {habit.current_streak > 0 && (
                <div className="flex items-center gap-1 text-sm text-orange-400 mb-2">
                  <Flame className="w-4 h-4" />
                  <span>{habit.current_streak} day streak</span>
                </div>
              )}

              <div className="flex items-center justify-between text-xs text-slate-400 mb-3">
                <span>Longest: {habit.longest_streak}d</span>
                <span>{habit.total_completions} total</span>
              </div>

              <button
                onClick={() => logMutation.mutate({ id: habit.id })}
                className="btn-secondary w-full text-xs"
                disabled={logMutation.isPending}
              >
                {logMutation.isPending ? <Loader2 className="w-3 h-3 animate-spin" /> : <Check className="w-3 h-3" />}
                Log Completion
              </button>
            </div>
          ))}
        </div>
      ) : (
        <div className="card text-center py-12">
          <Activity className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-slate-300 mb-2">No habits yet</h3>
          <p className="text-sm text-slate-400">Create your first habit to start tracking.</p>
        </div>
      )}
    </div>
  );
}
