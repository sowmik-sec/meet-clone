"use client";

import { Navigation } from '@/components/ui/Navigation';

export default function SettingsLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <div className="min-h-screen bg-gray-50">
            <Navigation />
            <main>
                {children}
            </main>
        </div>
    );
}
