'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { Calendar, Clock, LayoutDashboard, LogOut, Plus, Video } from 'lucide-react';

export function Navigation() {
    const pathname = usePathname();
    const router = useRouter();
    const { logout } = useAuth();

    const isActive = (path: string) => pathname === path;

    const handleLogout = async () => {
        await logout();
        router.push('/login');
    };

    return (
        <nav className="border-b bg-white">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex justify-between h-16">
                    <div className="flex">
                        <div className="flex-shrink-0 flex items-center cursor-pointer" onClick={() => router.push('/dashboard')}>
                            <Video className="h-8 w-8 text-blue-600" />
                            <span className="ml-2 text-xl font-bold text-gray-900">Meet Clone</span>
                        </div>
                        <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                            <Link
                                href="/dashboard"
                                className={cn(
                                    "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                    isActive('/dashboard')
                                        ? "border-blue-500 text-gray-900"
                                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                                )}
                            >
                                <LayoutDashboard className="w-4 h-4 mr-2" />
                                Dashboard
                            </Link>
                            <Link
                                href="/schedule"
                                className={cn(
                                    "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                    isActive('/schedule')
                                        ? "border-blue-500 text-gray-900"
                                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                                )}
                            >
                                <Calendar className="w-4 h-4 mr-2" />
                                Appointments
                            </Link>
                            <Link
                                href="/schedule/availability"
                                className={cn(
                                    "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                    isActive('/schedule/availability')
                                        ? "border-blue-500 text-gray-900"
                                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                                )}
                            >
                                <Clock className="w-4 h-4 mr-2" />
                                Availability
                            </Link>

                            <Link
                                href="/settings/event-types"
                                className={cn(
                                    "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                    isActive('/settings/event-types')
                                        ? "border-blue-500 text-gray-900"
                                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                                )}
                            >
                                <Clock className="w-4 h-4 mr-2" />
                                Event Types
                            </Link>

                            <Link
                                href="/settings/profile"
                                className={cn(
                                    "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                    isActive('/settings/profile')
                                        ? "border-blue-500 text-gray-900"
                                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                                )}
                            >
                                <Clock className="w-4 h-4 mr-2" />
                                Profile
                            </Link>
                        </div>
                    </div>
                    <div className="hidden sm:ml-6 sm:flex sm:items-center space-x-4">
                        <Button onClick={() => router.push('/schedule/new')} size="sm">
                            <Plus className="w-4 h-4 mr-2" />
                            New Appointment
                        </Button>
                        <Button variant="ghost" onClick={handleLogout} size="sm">
                            <LogOut className="w-4 h-4 mr-2" />
                            Logout
                        </Button>
                    </div>
                </div>
            </div>
        </nav >
    );
}
