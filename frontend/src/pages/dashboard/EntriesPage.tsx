import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { entriesApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { PenLine, Search, Filter, Plus, Trash2, Loader2, MoreHorizontal } from 'lucide-react';
import { format } from 'date-fns';
import toast from 'react-hot-toast';

export default function EntriesPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [entryType, setEntryType] = useState('');
  const [moodMin, setMoodMin] = useState('');
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['entries', page, search, entryType, moodMin],
    queryFn: () => entriesApi.list({ page, page_size: 20, search: search || undefined, entry_type: entryType || undefined, mood_min: moodMin ? parseInt(moodMin) : undefined }),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => entriesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      toast.success('Entry deleted');
    },
    onError: (err: any) => toast.error(err.message || 'Failed to delete'),
  });

  const getMoodColor = (score: number | null) => {
    if (!score) return 'text-slate-500';
    if (score <= 3) return 'text-red-400';
    if (score <= 6) return 'text-yellow-400';
    return 'text-green-400';
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Entries</h1>
          <p className="page-subtitle">Track your moods, activities, and journal entries</p>
        </div>
        <Link to="/entries/new" className="btn-primary">
          <Plus className="w-4 h-4" /> New Entry
        </Link>
      </div>

      <div className="flex flex-wrap gap-3 mb-6">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            placeholder="Search entries..."
            className="input pl-10"
          />
        </div>
        <select value={entryType} onChange={(e) => { setEntryType(e.target.value); setPage(1); }} className="input w-auto">
          <option value="">All Types</option>
          <option value="mood">Mood</option>
          <option value="activity">Activity</option>
          <option value="journal">Journal</option>
          <option value="sleep">Sleep</option>
          <option value="exercise">Exercise</option>
          <option value="weight">Weight</option>
          <option value="medication">Medication</option>
          <option value="custom">Custom</option>
        </select>
        <select value={moodMin} onChange={(e) => { setMoodMin(e.target.value); setPage(1); }} className="input w-auto">
          <option value="">Any Mood</option>
          <option value="7">Good (7+)</option>
          <option value="4">Okay (4+)</option>
          <option value="1">Low (1-3)</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
        </div>
      ) : data?.data && data.data.length > 0 ? (
        <>
          <div className="space-y-3">
            {data.data.map((entry) => (
              <div key={entry.id} className="card-hover flex items-start justify-between" onClick={() => navigate(`/entries/${entry.id}`)}>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="badge badge-primary text-xs">{entry.entry_type}</span>
                    {entry.mood_score && (
                      <span className={`text-sm font-medium ${getMoodColor(entry.mood_score)}`}>
                        {entry.mood_score}/10
                      </span>
                    )}
                    {entry.visibility === 'public' && <span className="badge badge-success text-xs">Public</span>}
                  </div>
                  <h3 className="text-sm font-medium text-slate-100 truncate">{entry.title || 'Untitled'}</h3>
                  {entry.description && <p className="text-xs text-slate-400 mt-1 line-clamp-1">{entry.description}</p>}
                  <p className="text-xs text-slate-500 mt-1">{format(new Date(entry.entry_date), 'MMM d, yyyy')}</p>
                </div>
                <button
                  onClick={(e) => { e.stopPropagation(); deleteMutation.mutate(entry.id); }}
                  className="p-2 text-slate-500 hover:text-danger transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>

          {data.total_pages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-6">
              <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary text-sm">
                Previous
              </button>
              <span className="text-sm text-slate-400">
                Page {page} of {data.total_pages}
              </span>
              <button onClick={() => setPage(p => Math.min(data.total_pages, p + 1))} disabled={page === data.total_pages} className="btn-secondary text-sm">
                Next
              </button>
            </div>
          )}
        </>
      ) : (
        <div className="card text-center py-12">
          <PenLine className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-slate-300 mb-2">No entries yet</h3>
          <p className="text-sm text-slate-400 mb-4">Start tracking your life by creating your first entry.</p>
          <Link to="/entries/new" className="btn-primary">Create Entry</Link>
        </div>
      )}
    </div>
  );
}
