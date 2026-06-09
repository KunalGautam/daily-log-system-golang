import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { publicApi } from '@/services/api';
import { Card } from '@/components/ui/card';
import { Globe, Loader2, Search, User } from 'lucide-react';
import { format } from 'date-fns';

export default function TimelinePage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['public-timeline', page, search],
    queryFn: () => publicApi.getTimeline({ page, search: search || undefined }),
  });

  const getMoodColor = (score: number | null) => {
    if (!score) return '';
    if (score <= 3) return 'text-red-400';
    if (score <= 6) return 'text-yellow-400';
    return 'text-green-400';
  };

  return (
    <div className="min-h-screen bg-slate-950">
      <header className="border-b border-slate-800">
        <div className="max-w-4xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Globe className="w-5 h-5 text-primary" />
            <h1 className="text-lg font-bold text-slate-100">Public Timeline</h1>
          </div>
          <a href="/" className="text-sm text-slate-400 hover:text-slate-100">Home</a>
        </div>
      </header>

      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            placeholder="Search public entries..."
            className="input pl-10"
          />
        </div>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
          </div>
        ) : data?.data && data.data.length > 0 ? (
          <>
            <div className="space-y-4">
              {data.data.map((entry) => (
                <div key={entry.id} className="card">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                      <User className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-slate-200">
                        {entry.user?.display_name || entry.user?.username || 'Anonymous'}
                      </p>
                      <p className="text-xs text-slate-500">
                        {format(new Date(entry.entry_date), 'MMM d, yyyy')} &middot; {entry.entry_type}
                      </p>
                    </div>
                    {entry.mood_score && (
                      <span className={`ml-auto text-lg font-bold ${getMoodColor(entry.mood_score)}`}>
                        {entry.mood_score}/10
                      </span>
                    )}
                  </div>
                  {entry.title && <h3 className="text-base font-semibold text-slate-100 mb-1">{entry.title}</h3>}
                  {entry.description && <p className="text-sm text-slate-300">{entry.description}</p>}
                  {entry.tags && entry.tags.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-3">
                      {entry.tags.map((tag) => (
                        <span key={tag.id} className="badge badge-primary text-xs">{tag.name}</span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {data.total_pages > 1 && (
              <div className="flex items-center justify-center gap-2 mt-6">
                <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary text-sm">Previous</button>
                <span className="text-sm text-slate-400">Page {page} of {data.total_pages}</span>
                <button onClick={() => setPage(p => Math.min(data.total_pages, p + 1))} disabled={page === data.total_pages} className="btn-secondary text-sm">Next</button>
              </div>
            )}
          </>
        ) : (
          <div className="card text-center py-12">
            <Globe className="w-12 h-12 text-slate-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-slate-300 mb-2">No public entries yet</h3>
            <p className="text-sm text-slate-400">Be the first to share your life moments.</p>
          </div>
        )}
      </div>
    </div>
  );
}
