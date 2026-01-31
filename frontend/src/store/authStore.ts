import { create } from 'zustand';
import { User } from '@/types/auth';
import { cookies, AUTH_COOKIE_NAME, USER_COOKIE_NAME } from '@/lib/cookies';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  hasHydrated: boolean;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
  initializeAuth: () => void;
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  hasHydrated: false,
  setAuth: (user, token) => {
    // Store in cookies (7 days expiry)
    cookies.set(AUTH_COOKIE_NAME, token, { days: 7 });
    cookies.set(USER_COOKIE_NAME, JSON.stringify(user), { days: 7 });
    set({ user, token, isAuthenticated: true });
  },
  clearAuth: () => {
    // Remove from cookies
    cookies.remove(AUTH_COOKIE_NAME);
    cookies.remove(USER_COOKIE_NAME);
    set({ user: null, token: null, isAuthenticated: false });
  },
  initializeAuth: () => {
    // Initialize auth from cookies on mount
    if (typeof window === 'undefined') {
      set({ hasHydrated: true });
      return;
    }

    const token = cookies.get(AUTH_COOKIE_NAME);
    const userStr = cookies.get(USER_COOKIE_NAME);
    
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr) as User;
        set({ user, token, isAuthenticated: true, hasHydrated: true });
      } catch {
        // Invalid user data, clear everything
        cookies.remove(AUTH_COOKIE_NAME);
        cookies.remove(USER_COOKIE_NAME);
        set({ user: null, token: null, isAuthenticated: false, hasHydrated: true });
      }
    } else {
      set({ hasHydrated: true });
    }
  },
}));
