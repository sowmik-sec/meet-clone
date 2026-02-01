'use client';

import { useAuthStore } from '@/store/authStore';
import { useEffect, useRef } from 'react';

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const initializeAuth = useAuthStore((state) => state.initializeAuth);
    const initialized = useRef(false);

    useEffect(() => {
        if (!initialized.current) {
            initializeAuth();
            initialized.current = true;
        }
    }, [initializeAuth]);

    return <>{children}</>;
}
