import { useAuthStore } from '@/store/authStore';
import { authApi } from '@/lib/api/auth';
import { LoginRequest, RegisterRequest } from '@/types/auth';
import { useState } from 'react';

export function useAuth() {
  const { user, isAuthenticated, hasHydrated, setAuth, clearAuth, initializeAuth } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const login = async (data: LoginRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await authApi.login(data);
      setAuth(response.user);
      return response;
    } catch (error: unknown) {
      const message = error && typeof error === 'object' && 'response' in error
        ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Login failed'
        : 'Login failed';
      setError(message);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (data: RegisterRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await authApi.register(data);
      setAuth(response.user);
      return response;
    } catch (error: unknown) {
      const message = error && typeof error === 'object' && 'response' in error
        ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Registration failed'
        : 'Registration failed';
      setError(message);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } catch (error) {
      console.error('Logout failed', error);
    } finally {
      clearAuth();
    }
  };

  return {
    user,
    isAuthenticated,
    hasHydrated,
    isLoading,
    error,
    login,
    register,
    logout,
    initializeAuth,
  };
}
