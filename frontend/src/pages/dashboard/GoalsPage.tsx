import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { goalsApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Target, Plus, Loader2, Trash2, TrendingUp } from 'lucide-react';
import { useForm } from 'react-hook-form';
import toast from 'react-hot-toast';

export default function GoalsPage() {
  const [showForm, setShowForm] = useState(false);
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['goals'],
    queryFn: () => goalsApi.list({ page_size: 50 }),
  });

  const { data: stats } = useQuery({
    queryKey: ['goals-stats'],
    queryFn: () => goalsApi.getStats(),
  });

  const { register, handleSubmit, reset, formState: { isSubmitting } } = useForm();

  const createMutation = useMutation({
    mutationFn: (data: any) => goalsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['goals-stats'] });
      toast.success('Goal created');
      setShowForm(false);
      reset();
    },
    onError: (err: any) => toast.error(err.message || 'Failed to create'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => goalsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['goals-stats'] });
      toast.success('Goal deleted');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to delete'),
  });

  const addProgressMutation = useMutation({
    mutationFn: ({ id, value }: { id: string; value: number }) => goalsApi.addProgress(id, { value }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['goals-stats'] });
      toast.success('Progress updated');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to update'),
  });

  const [progressValues, setProgressValues] = useState<Record<string, string>>({});

  const onSubmit = (formData: any) => {
    createMutation.mutate({
      title: formData.title,
      description: formData.description,
      goal_type: formData.goal_type || 'daily',
      target_value: parseFloat(formData.target_value) || 1,
      unit: formData.unit,
    });
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Goals</h1>
          <p className="page-subtitle">Set and track your personal goals</p>
        </div>
        <button onClick={() => setShowForm(!showForm)} className="btn-primary">
          <Plus className="w-4 h-4" /> New Goal
        </button>
      </div>

      {stats && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="card">
            <p className="text-2xl font-bold text-slate-100">{stats.total_goals ?? 0}</p>
            <p className="text-xs text-slate-400">Total Goals</p>
          </div>
          <div className="card">
            <p className="text-2xl font-bold text-green-400">{stats.active_goals ?? 0}</p>
            <p className="text-xs text-slate-400">Active</p>
          </div>
          <div className="card">
            <p className="text-2xl font-bold text-blue-400">{stats.completed_goals ?? 0}</p>
            <p className="text-xs text-slate-400">Completed</p>
          </div>
        </div>
      )}

      {showForm && (
        <div className="card mb-6">
          <h3 className="text-lg font-semibold text-slate-100 mb-4">New Goal</h3>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="flex gap-4">
              <div className="flex-1">
                <label className="label">Title</label>
                <input {...register('title', { required: true })} className="input" placeholder="Read 50 books this year" />
              </div>
              <div className="w-40">
                <label className="label">Type</label>
                <select {...register('goal_type')} className="input">
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                  <option value="yearly">Yearly</option>
                </select>
              </div>
            </div>
            <div className="flex gap-4">
              <div className="w-40">
                <label className="label">Target Value</label>
                <input type="number" {...register('target_value')} className="input" step="any" />
              </div>
              <div className="w-40">
                <label className="label">Unit</label>
                <input {...register('unit')} className="input" placeholder="books, kg, km" />
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
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {data.data.map((goal) => {
            const progress = goal.target_value > 0 ? (goal.current_value / goal.target_value) * 100 : 0;
            return (
              <div key={goal.id} className="card">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-secondary/10 flex items-center justify-center">
                      <Target className="w-5 h-5 text-secondary" />
                    </div>
                    <div>
                      <h3 className="text-sm font-medium text-slate-100">{goal.title}</h3>
                      <p className="text-xs text-slate-400">{goal.goal_type}{goal.unit ? ` (${goal.unit})` : ''}</p>
                    </div>
                  </div>
                  <button onClick={() => deleteMutation.mutate(goal.id)} className="text-slate-500 hover:text-danger transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>

                <div className="mb-3">
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-slate-300">{goal.current_value}</span>
                    <span className="text-slate-400">{goal.target_value}</span>
                  </div>
                  <div className="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden">
                    <div className="h-full bg-secondary rounded-full transition-all" style={{ width: `${Math.min(progress, 100)}%` }} />
                  </div>
                  <p className="text-xs text-slate-400 mt-1">{Math.round(progress)}% complete</p>
                </div>

                {goal.is_completed && (
                  <div className="badge badge-success text-xs mb-3">Completed!</div>
                )}

                <div className="flex gap-2">
                  <input
                    type="number"
                    value={progressValues[goal.id] || ''}
                    onChange={(e) => setProgressValues(prev => ({ ...prev, [goal.id]: e.target.value }))}
                    className="input flex-1 text-sm"
                    placeholder="Add progress..."
                    step="any"
                  />
                  <button
                    onClick={() => {
                      const val = parseFloat(progressValues[goal.id] || '0');
                      if (val > 0) {
                        addProgressMutation.mutate({ id: goal.id, value: val });
                        setProgressValues(prev => ({ ...prev, [goal.id]: '' }));
                      }
                    }}
                    className="btn-secondary text-sm"
                    disabled={addProgressMutation.isPending}
                  >
                    <TrendingUp className="w-4 h-4" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="card text-center py-12">
          <Target className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-slate-300 mb-2">No goals yet</h3>
          <p className="text-sm text-slate-400">Set your first goal to start tracking progress.</p>
        </div>
      )}
    </div>
  );
}
