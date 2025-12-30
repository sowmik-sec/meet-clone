import { useAuthStore } from '@/store/authStore';
import { authApi } from '@/lib/api/auth';
import { LoginRequest, RegisterRequest } from '@/types/auth';
import { useState } from 'react';

export function useAuth() {
  const { user, token, isAuthenticated, hasHydrated, setAuth, clearAuth } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const login = async (data: LoginRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await authApi.login(data);
      setAuth(response.user, response.token);
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
      setAuth(response.user, response.token);
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

  const logout = () => {
    clearAuth();
  };

  return {
    user,
    token,
    isAuthenticated,
    hasHydrated,
    isLoading,
    error,
    login,
    register,
    logout,
  };
}
