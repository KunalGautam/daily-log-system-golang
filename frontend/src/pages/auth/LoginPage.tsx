import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import toast from 'react-hot-toast';
import { authApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { LogIn, Eye, EyeOff, Loader2 } from 'lucide-react';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
  totp_code: z.string().optional(),
});

type LoginForm = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);
  const [showTOTP, setShowTOTP] = useState(false);

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginForm) => {
    try {
      const response = await authApi.login(data);
      setAuth(response.user, response.access_token, response.refresh_token);
      toast.success('Welcome back!');
      navigate('/dashboard');
    } catch (err: any) {
      if (err.status === 401 && err.message?.includes('TOTP')) {
        setShowTOTP(true);
        toast.error('TOTP code required');
      } else {
        toast.error(err.message || 'Login failed');
      }
    }
  };

  return (
    <div className="w-full max-w-md">
      <div className="card">
        <h1 className="text-2xl font-bold text-slate-100 mb-2">Sign In</h1>
        <p className="text-slate-400 text-sm mb-6">Welcome back to LifeLog</p>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="label">Email</label>
            <input type="email" {...register('email')} className={errors.email ? 'input-error' : 'input'} placeholder="you@example.com" autoComplete="email" />
            {errors.email && <p className="text-danger text-xs mt-1">{errors.email.message}</p>}
          </div>

          <div>
            <label className="label">Password</label>
            <div className="relative">
              <input type={showPassword ? 'text' : 'password'} {...register('password')} className={errors.password ? 'input-error pr-10' : 'input pr-10'} placeholder="Your password" autoComplete="current-password" />
              <button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-100">
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            {errors.password && <p className="text-danger text-xs mt-1">{errors.password.message}</p>}
          </div>

          {showTOTP && (
            <div>
              <label className="label">TOTP Code</label>
              <input type="text" {...register('totp_code')} className="input" placeholder="000000" maxLength={6} />
            </div>
          )}

          <button type="submit" disabled={isSubmitting} className="btn-primary w-full">
            {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <LogIn className="w-4 h-4" />}
            Sign In
          </button>
        </form>

        <div className="mt-4 text-center space-y-2">
          <Link to="/forgot-password" className="text-sm text-primary hover:text-primary-dark transition-colors">Forgot password?</Link>
        </div>
        <p className="mt-4 text-center text-sm text-slate-400">
          Don't have an account?{' '}
          <Link to="/register" className="text-primary hover:text-primary-dark transition-colors">Sign up</Link>
        </p>
      </div>
    </div>
  );
}
