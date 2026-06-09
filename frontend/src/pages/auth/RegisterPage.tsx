import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import toast from 'react-hot-toast';
import { authApi } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { UserPlus, Loader2 } from 'lucide-react';

const registerSchema = z.object({
  email: z.string().email('Invalid email address'),
  username: z.string().min(3, 'Username must be at least 3 characters').max(30, 'Username too long').regex(/^[a-zA-Z0-9_]+$/, 'Only letters, numbers, and underscores'),
  password: z.string().min(8, 'Password must be at least 8 characters').regex(/[A-Z]/, 'Must contain an uppercase letter').regex(/[0-9]/, 'Must contain a number'),
  display_name: z.string().max(50).optional(),
});

type RegisterForm = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterForm) => {
    try {
      const response = await authApi.register(data);
      toast.success('Account created!');
      navigate('/dashboard');
    } catch (err: any) {
      toast.error(err.message || 'Registration failed');
    }
  };

  return (
    <div className="w-full max-w-md">
      <div className="card">
        <h1 className="text-2xl font-bold text-slate-100 mb-2">Create Account</h1>
        <p className="text-slate-400 text-sm mb-6">Start tracking your life</p>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="label">Email</label>
            <input type="email" {...register('email')} className={errors.email ? 'input-error' : 'input'} placeholder="you@example.com" />
            {errors.email && <p className="text-danger text-xs mt-1">{errors.email.message}</p>}
          </div>

          <div>
            <label className="label">Username</label>
            <input type="text" {...register('username')} className={errors.username ? 'input-error' : 'input'} placeholder="yourusername" />
            {errors.username && <p className="text-danger text-xs mt-1">{errors.username.message}</p>}
          </div>

          <div>
            <label className="label">Display Name</label>
            <input type="text" {...register('display_name')} className="input" placeholder="Your Name (optional)" />
          </div>

          <div>
            <label className="label">Password</label>
            <input type="password" {...register('password')} className={errors.password ? 'input-error' : 'input'} placeholder="At least 8 characters" />
            {errors.password && <p className="text-danger text-xs mt-1">{errors.password.message}</p>}
          </div>

          <button type="submit" disabled={isSubmitting} className="btn-primary w-full">
            {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <UserPlus className="w-4 h-4" />}
            Create Account
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-slate-400">
          Already have an account?{' '}
          <Link to="/login" className="text-primary hover:text-primary-dark transition-colors">Sign in</Link>
        </p>
      </div>
    </div>
  );
}
