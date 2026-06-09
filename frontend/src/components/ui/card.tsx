import { ReactNode } from 'react';
import { clsx } from 'clsx';

export function Card({ className, children, ...props }: { className?: string; children: ReactNode; [key: string]: any }) {
  return <div className={clsx('rounded-xl border border-slate-800 bg-slate-900 p-6', className)} {...props}>{children}</div>;
}

export function CardHeader({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={clsx('mb-4', className)}>{children}</div>;
}

export function CardTitle({ className, children }: { className?: string; children: ReactNode }) {
  return <h3 className={clsx('text-lg font-semibold text-slate-100', className)}>{children}</h3>;
}

export function CardContent({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={className}>{children}</div>;
}
