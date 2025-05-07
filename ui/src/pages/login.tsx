import React from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from '@tanstack/react-router';
import { useAuthStore } from '@/store/auth-store';
import { Button } from '@/components/ui/button';
import { LuServer } from 'react-icons/lu';
import { useTheme } from '@/contexts/theme-context';

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export const LoginPage: React.FC = () => {
  const { theme, toggleTheme } = useTheme();
  const { login, isLoading, error, clearError } = useAuthStore();
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormValues) => {
    clearError();
    try {
      await login(data.username, data.password);
      // Only navigate if login was successful
      navigate({ to: '/' });
    } catch (error) {
      // Error is already handled in the store
      console.error('Login failed:', error);
    }
  };

  return (
    <div className="min-h-screen flex flex-col justify-center items-center p-4 bg-background">
      <div className="w-full max-w-md">
        <div className="flex justify-center mb-8">
          <LuServer className="h-12 w-12 text-primary" />
        </div>
        
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold">LibGo KVM</h1>
          <p className="text-muted-foreground">Sign in to manage your virtual machines</p>
        </div>
        
        <div className="bg-card p-8 rounded-lg shadow-lg border">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-2 rounded">
                {error}
              </div>
            )}
            
            <div>
              <label htmlFor="username" className="block text-sm font-medium mb-1">
                Username
              </label>
              <input
                id="username"
                type="text"
                {...register('username')}
                className="w-full p-2 rounded border bg-background"
                disabled={isLoading}
              />
              {errors.username && (
                <p className="text-red-500 text-xs mt-1">{errors.username.message}</p>
              )}
            </div>
            
            <div>
              <label htmlFor="password" className="block text-sm font-medium mb-1">
                Password
              </label>
              <input
                id="password"
                type="password"
                {...register('password')}
                className="w-full p-2 rounded border bg-background"
                disabled={isLoading}
              />
              {errors.password && (
                <p className="text-red-500 text-xs mt-1">{errors.password.message}</p>
              )}
            </div>
            
            <Button
              type="submit"
              className="w-full"
              disabled={isLoading}
            >
              {isLoading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
        </div>
        
        <div className="mt-4 text-center">
          <button 
            onClick={toggleTheme}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Switch to {theme === 'dark' ? 'light' : 'dark'} mode
          </button>
        </div>
      </div>
    </div>
  );
};