import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import toast from 'react-hot-toast';
import { authApi } from '@/services/api';
import { Mail, Loader2, ArrowLeft, CheckCircle } from 'lucide-react';

const emailSchema = z.object({ email: z.string().email('Invalid email') });
type EmailForm = z.infer<typeof emailSchema>;

export default function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<EmailForm>({ resolver: zodResolver(emailSchema) });

  const onSubmit = async (data: EmailForm) => {
    try {
      await authApi.forgotPassword(data.email);
      setSent(true);
      toast.success('Reset link sent if email exists');
    } catch (err: any) {
      toast.error(err.message || 'Failed');
    }
  };

  if (sent) {
    return (
      <div className="w-full max-w-md">
        <div className="card text-center">
          <CheckCircle className="w-12 h-12 text-accent mx-auto mb-4" />
          <h1 className="text-xl font-bold text-slate-100 mb-2">Check Your Email</h1>
          <p className="text-slate-400 text-sm mb-6">If an account exists, we've sent a password reset link.</p>
          <Link to="/login" className="text-primary hover:text-primary-dark text-sm flex items-center justify-center gap-2">
            <ArrowLeft className="w-4 h-4" /> Back to sign in
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md">
      <div className="card">
        <h1 className="text-2xl font-bold text-slate-100 mb-2">Reset Password</h1>
        <p className="text-slate-400 text-sm mb-6">Enter your email to receive a reset link.</p>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="label">Email</label>
            <input type="email" {...register('email')} className={errors.email ? 'input-error' : 'input'} placeholder="you@example.com" />
            {errors.email && <p className="text-danger text-xs mt-1">{errors.email.message}</p>}
          </div>
          <button type="submit" disabled={isSubmitting} className="btn-primary w-full">
            {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Mail className="w-4 h-4" />}
            Send Reset Link
          </button>
        </form>
        <Link to="/login" className="mt-4 block text-center text-sm text-primary hover:text-primary-dark transition-colors">
          Back to sign in
        </Link>
      </div>
    </div>
  );
}
