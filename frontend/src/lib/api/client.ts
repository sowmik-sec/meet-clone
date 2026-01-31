import axios, { AxiosError } from 'axios';
import { cookies, AUTH_COOKIE_NAME } from '@/lib/cookies';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: `${API_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token from cookies
api.interceptors.request.use(
  (config) => {
    const token = cookies.get(AUTH_COOKIE_NAME);
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Clear cookies
      cookies.remove(AUTH_COOKIE_NAME);
      cookies.remove('auth_user');
      
      // Get current path for redirect after login
      const currentPath = window.location.pathname;
      const redirectParam = currentPath !== '/login' && currentPath !== '/register' 
        ? `?redirect=${encodeURIComponent(currentPath)}` 
        : '';
      
      window.location.href = `/login${redirectParam}`;
    }
    return Promise.reject(error);
  }
);

export default api;
