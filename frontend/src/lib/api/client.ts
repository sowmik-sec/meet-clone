import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { cookies, AUTH_COOKIE_NAME } from '@/lib/cookies';

// Extend axios config to support custom properties
interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
  skipRedirect?: boolean;
}

// Use relative path for proxying in development/production
const API_URL = '';

export const api = axios.create({
  baseURL: `${API_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

// Request interceptor to add auth token from cookies
// Request interceptor (Simplified as cookies are handled automatically)
api.interceptors.request.use(
  (config) => {
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
    const config = error.config as CustomAxiosRequestConfig;

    if (error.response?.status === 401 && !config?.skipRedirect) {
      // Clear local user data
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
