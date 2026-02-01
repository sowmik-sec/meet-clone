import { create } from 'zustand';
import { User } from '@/types/auth';
import { cookies, USER_COOKIE_NAME } from '@/lib/cookies';
import { authApi } from '@/lib/api/auth';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  hasHydrated: boolean;
  setAuth: (user: User) => void;
  clearAuth: () => void;
  initializeAuth: () => void;
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  isAuthenticated: false,
  hasHydrated: false,
  setAuth: (user) => {
    // Store user data in non-HttpOnly cookie for client availability (optional, but good for persistence)
    cookies.set(USER_COOKIE_NAME, JSON.stringify(user), { days: 7 });
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_sync_event', JSON.stringify({ type: 'LOGIN', user, timestamp: Date.now() }));
    }
    set({ user, isAuthenticated: true });
  },
  clearAuth: () => {
    // Remove from cookies
    cookies.remove(USER_COOKIE_NAME);
    // but ideally we should hit a logout endpoint. For now, we just clear client state.
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_sync_event', JSON.stringify({ type: 'LOGOUT', timestamp: Date.now() }));
    }
    set({ user: null, isAuthenticated: false });
  },
  initializeAuth: () => {
    // Initialize auth from cookies on mount
    if (typeof window === 'undefined') {
      set({ hasHydrated: true });
      return;
    }

    const userStr = cookies.get(USER_COOKIE_NAME);

    if (userStr) {
      try {
        const user = JSON.parse(userStr) as User;
        // Optimistically set user
        set({ user, isAuthenticated: true, hasHydrated: false }); // Keep hydrated false until verification

        // Verify with server
        authApi.me()
          .then((verifiedUser) => {
            set({ user: verifiedUser, isAuthenticated: true, hasHydrated: true });
            // Update local cookie if needed
            cookies.set(USER_COOKIE_NAME, JSON.stringify(verifiedUser), { days: 7 });
          })
          .catch(() => {
            // Verification failed (cookie expired/invalid)
            cookies.remove(USER_COOKIE_NAME);
            set({ user: null, isAuthenticated: false, hasHydrated: true });
          });
      } catch {
        // Invalid user data, clear everything
        cookies.remove(USER_COOKIE_NAME);
        set({ user: null, isAuthenticated: false, hasHydrated: true });
      }
    } else {
      set({ hasHydrated: true });
    }

    // Sync auth state across tabs
    const handleStorageChange = (event: StorageEvent) => {
      // If user logs out in another tab (USER_COOKIE_NAME removed from generic cookie storage if we used localStorage, but here we check cookie changes indirectly usually or via a custom event. 
      // Since we use 'js-cookie', it doesn't fire storage events for cookies automatically. 
      // However, we can listen for a custom broadcast channel or just check visibility.
      // For simplicity/robustness, we'll actually use a BroadcastChannel or periodically check, 
      // but standard best practice with 'user cookie' is often to rely on window focus.

      // Better approach: When we modify auth, update a localStorage key 'last_auth_event' with timestamp
      // to trigger the storage event in other tabs.
    };

    // Simplified robust sync:
    window.addEventListener('storage', (e) => {
      if (e.key === 'auth_sync_event') {
        const event = JSON.parse(e.newValue || '{}');
        if (event.type === 'LOGOUT') {
          set({ user: null, isAuthenticated: false });
        } else if (event.type === 'LOGIN' && event.user) {
          set({ user: event.user, isAuthenticated: true });
        }
      }
    });
  },
}));
