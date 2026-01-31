/**
 * Cookie utility functions for authentication
 * Using cookies is more secure than localStorage for auth tokens
 */

export const AUTH_COOKIE_NAME = 'auth_token';
export const USER_COOKIE_NAME = 'auth_user';

interface CookieOptions {
  days?: number;
  path?: string;
  secure?: boolean;
  sameSite?: 'strict' | 'lax' | 'none';
}

export const cookies = {
  set(name: string, value: string, options: CookieOptions = {}) {
    const {
      days = 7,
      path = '/',
      secure = process.env.NODE_ENV === 'production',
      sameSite = 'lax',
    } = options;

    let cookieString = `${encodeURIComponent(name)}=${encodeURIComponent(value)}`;
    
    if (days) {
      const date = new Date();
      date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
      cookieString += `; expires=${date.toUTCString()}`;
    }
    
    cookieString += `; path=${path}`;
    
    if (secure) {
      cookieString += '; secure';
    }
    
    cookieString += `; SameSite=${sameSite}`;
    
    document.cookie = cookieString;
  },

  get(name: string): string | null {
    if (typeof document === 'undefined') return null;
    
    const nameEQ = encodeURIComponent(name) + '=';
    const cookies = document.cookie.split(';');
    
    for (let i = 0; i < cookies.length; i++) {
      let cookie = cookies[i];
      while (cookie.charAt(0) === ' ') {
        cookie = cookie.substring(1, cookie.length);
      }
      if (cookie.indexOf(nameEQ) === 0) {
        return decodeURIComponent(cookie.substring(nameEQ.length, cookie.length));
      }
    }
    
    return null;
  },

  remove(name: string, path: string = '/') {
    document.cookie = `${encodeURIComponent(name)}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=${path};`;
  },

  // Helper to check if cookies are available
  isAvailable(): boolean {
    if (typeof document === 'undefined') return false;
    try {
      const testKey = '__cookie_test__';
      cookies.set(testKey, 'test', { days: 1 });
      const result = cookies.get(testKey) === 'test';
      cookies.remove(testKey);
      return result;
    } catch {
      return false;
    }
  },
};
