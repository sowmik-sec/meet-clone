'use client';

import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { useEffect, useState } from 'react';

export function HeaderAuthButtons() {
    const { isAuthenticated, logout, hasHydrated } = useAuth();
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
    }, []);

    // Prevent hydration mismatch
    if (!mounted || !hasHydrated) {
        return (
            <div className="flex gap-4">
                <Button variant="ghost" disabled>
                    Loading...
                </Button>
            </div>
        );
    }

    if (isAuthenticated) {
        return (
            <div className="flex gap-4">
                <Link href="/dashboard">
                    <Button variant="ghost">
                        Dashboard
                    </Button>
                </Link>
                <Button onClick={logout}>
                    Logout
                </Button>
            </div>
        );
    }

    return (
        <div className="flex gap-4">
            <Link href="/login">
                <Button variant="ghost">
                    Login
                </Button>
            </Link>
            <Link href="/register">
                <Button>
                    Get Started
                </Button>
            </Link>
        </div>
    );
}
