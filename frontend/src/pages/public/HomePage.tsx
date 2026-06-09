import { Link } from 'react-router-dom';
import { Activity, Shield, BarChart3, Globe, Github, ArrowRight } from 'lucide-react';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-slate-950">
      <header className="border-b border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
              <Activity className="w-5 h-5 text-white" />
            </div>
            <span className="text-lg font-bold text-slate-100">LifeLog</span>
          </div>
          <div className="flex items-center gap-4">
            <Link to="/timeline" className="text-sm text-slate-400 hover:text-slate-100 transition-colors">Timeline</Link>
            <Link to="/stats" className="text-sm text-slate-400 hover:text-slate-100 transition-colors">Stats</Link>
            <Link to="/login" className="btn-secondary text-sm">Sign In</Link>
            <Link to="/register" className="btn-primary text-sm">Get Started</Link>
          </div>
        </div>
      </header>

      <main>
        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 text-center">
          <h1 className="text-5xl font-bold text-slate-100 mb-6">
            Track Your Life,<br />
            <span className="text-primary">Understand Yourself</span>
          </h1>
          <p className="text-xl text-slate-400 max-w-2xl mx-auto mb-10">
            A self-hosted life logging and analytics platform. Track moods, habits, goals, 
            and more. Gain insights into your daily patterns.
          </p>
          <div className="flex items-center justify-center gap-4">
            <Link to="/register" className="btn-primary text-lg px-8 py-3">
              Start Logging <ArrowRight className="w-5 h-5" />
            </Link>
            <Link to="/timeline" className="btn-secondary text-lg px-8 py-3">
              View Timeline
            </Link>
          </div>
        </section>

        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="card text-center">
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <Activity className="w-6 h-6 text-primary" />
              </div>
              <h3 className="text-lg font-semibold text-slate-100 mb-2">Mood & Activity Tracking</h3>
              <p className="text-sm text-slate-400">Log your moods, activities, journal entries, and more with rich markdown support.</p>
            </div>
            <div className="card text-center">
              <div className="w-12 h-12 rounded-xl bg-secondary/10 flex items-center justify-center mx-auto mb-4">
                <BarChart3 className="w-6 h-6 text-secondary" />
              </div>
              <h3 className="text-lg font-semibold text-slate-100 mb-2">Analytics & Insights</h3>
              <p className="text-sm text-slate-400">Visualize trends, discover patterns, and track your progress over time.</p>
            </div>
            <div className="card text-center">
              <div className="w-12 h-12 rounded-xl bg-accent/10 flex items-center justify-center mx-auto mb-4">
                <Shield className="w-6 h-6 text-accent" />
              </div>
              <h3 className="text-lg font-semibold text-slate-100 mb-2">Private & Secure</h3>
              <p className="text-sm text-slate-400">Self-hosted with end-to-end encryption. Your data stays yours.</p>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-slate-800 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between">
          <p className="text-sm text-slate-500">LifeLog &mdash; Self-hosted life logging platform</p>
          <a href="https://github.com" className="text-slate-400 hover:text-slate-100 transition-colors">
            <Github className="w-5 h-5" />
          </a>
        </div>
      </footer>
    </div>
  );
}
