import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { entriesApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Loader2, Save, ArrowLeft, Trash2 } from 'lucide-react';
import toast from 'react-hot-toast';
import { format } from 'date-fns';

interface EntryForm {
  title: string;
  description: string;
  markdown_content: string;
  entry_type: string;
  mood_score: number;
  activities: string;
  visibility: string;
  entry_date: string;
}

export default function EntryDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isNew = id === 'new';

  const { data: entry, isLoading } = useQuery({
    queryKey: ['entry', id],
    queryFn: () => entriesApi.get(id!),
    enabled: !isNew && !!id,
  });

  const { register, handleSubmit, reset, formState: { isSubmitting } } = useForm<EntryForm>();

  useEffect(() => {
    if (entry) {
      reset({
        title: entry.title || '',
        description: entry.description || '',
        markdown_content: entry.markdown_content || '',
        entry_type: entry.entry_type,
        mood_score: entry.mood_score || 0,
        activities: entry.activities || '',
        visibility: entry.visibility,
        entry_date: entry.entry_date?.split('T')[0] || format(new Date(), 'yyyy-MM-dd'),
      });
    }
  }, [entry, reset]);

  const createMutation = useMutation({
    mutationFn: (data: EntryForm) => entriesApi.create(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      toast.success('Entry created');
      navigate(`/entries/${result.id}`);
    },
    onError: (err: any) => toast.error(err.message || 'Failed to create'),
  });

  const updateMutation = useMutation({
    mutationFn: (data: EntryForm) => entriesApi.update(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      queryClient.invalidateQueries({ queryKey: ['entry', id] });
      toast.success('Entry updated');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to update'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => entriesApi.delete(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      toast.success('Entry deleted');
      navigate('/entries');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to delete'),
  });

  const onSubmit = (data: EntryForm) => {
    if (isNew) createMutation.mutate(data);
    else updateMutation.mutate(data);
  };

  if (!isNew && isLoading) {
    return (
      <div className="page-container flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="page-container max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <button onClick={() => navigate('/entries')} className="btn-ghost">
          <ArrowLeft className="w-4 h-4" /> Back
        </button>
        {!isNew && (
          <button onClick={() => deleteMutation.mutate()} className="btn-danger">
            <Trash2 className="w-4 h-4" /> Delete
          </button>
        )}
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="flex gap-4">
          <div className="flex-1">
            <label className="label">Title</label>
            <input {...register('title')} className="input" placeholder="How was your day?" />
          </div>
          <div className="w-40">
            <label className="label">Type</label>
            <select {...register('entry_type')} className="input" required>
              <option value="mood">Mood</option>
              <option value="activity">Activity</option>
              <option value="journal">Journal</option>
              <option value="sleep">Sleep</option>
              <option value="exercise">Exercise</option>
              <option value="weight">Weight</option>
              <option value="medication">Medication</option>
              <option value="custom">Custom</option>
            </select>
          </div>
        </div>

        <div className="flex gap-4">
          <div className="w-32">
            <label className="label">Mood (1-10)</label>
            <input type="number" {...register('mood_score', { valueAsNumber: true })} className="input" min={1} max={10} />
          </div>
          <div className="flex-1">
            <label className="label">Visibility</label>
            <select {...register('visibility')} className="input">
              <option value="private">Private</option>
              <option value="public">Public</option>
              <option value="unlisted">Unlisted</option>
            </select>
          </div>
          <div className="w-44">
            <label className="label">Date</label>
            <input type="date" {...register('entry_date')} className="input" />
          </div>
        </div>

        <div>
          <label className="label">Description</label>
          <input {...register('description')} className="input" placeholder="Brief summary..." />
        </div>

        <div>
          <label className="label">Activities (comma separated)</label>
          <input {...register('activities')} className="input" placeholder="reading, exercise, cooking" />
        </div>

        <div>
          <label className="label">Notes (Markdown)</label>
          <textarea {...register('markdown_content')} className="input min-h-[200px] font-mono" placeholder="Write your thoughts in markdown..." />
        </div>

        <button type="submit" disabled={isSubmitting} className="btn-primary">
          {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          {isNew ? 'Create Entry' : 'Save Changes'}
        </button>
      </form>
    </div>
  );
}
