'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { roomApi } from '@/lib/api/room';
import { Room } from '@/types/room';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';

export default function DashboardPage() {
  const router = useRouter();
  const { user, isAuthenticated, hasHydrated, logout } = useAuth();
  const [roomId, setRoomId] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [activeRooms, setActiveRooms] = useState<Room[]>([]);
  const [isLoadingRooms, setIsLoadingRooms] = useState(true);

  useEffect(() => {
    if (hasHydrated && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, hasHydrated, router]);

  useEffect(() => {
    const loadActiveRooms = async () => {
      try {
        setIsLoadingRooms(true);
        // Get user's active rooms from backend
        const rooms = await roomApi.getUserRooms();
        setActiveRooms(rooms);
      } catch (error) {
        console.error('Failed to load rooms:', error);
      } finally {
        setIsLoadingRooms(false);
      }
    };

    if (isAuthenticated) {
      loadActiveRooms();
    }
  }, [isAuthenticated]);

  const handleCreateRoom = async () => {
    setIsCreating(true);
    try {
      const room = await roomApi.createRoom();
      router.push(`/room/${room.id}`);
    } catch (err) {
      console.error('Failed to create room:', err);
    } finally {
      setIsCreating(false);
    }
  };

  const handleJoinRoom = (e: React.FormEvent) => {
    e.preventDefault();
    if (roomId.trim()) {
      router.push(`/room/${roomId.trim()}`);
    }
  };

  if (!hasHydrated || !isAuthenticated || !user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
      <nav className="border-b bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-2xl font-bold text-blue-600">Meet Clone</h1>
            <div className="flex items-center gap-4">
              <span className="text-gray-700">Welcome, {user.name}</span>
              <Button variant="ghost" onClick={logout}>
                Logout
              </Button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="grid md:grid-cols-2 gap-8">
          <Card>
            <CardHeader>
              <CardTitle>Create New Meeting</CardTitle>
              <CardDescription>
                Start a new meeting and invite participants
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button
                className="w-full"
                onClick={handleCreateRoom}
                disabled={isCreating}
              >
                {isCreating ? 'Creating...' : 'New Meeting'}
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Join Meeting</CardTitle>
              <CardDescription>
                Enter a meeting ID to join
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleJoinRoom} className="space-y-4">
                <Input
                  placeholder="Enter meeting ID"
                  value={roomId}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => setRoomId(e.target.value)}
                  required
                />
                <Button type="submit" variant="secondary" className="w-full">
                  Join
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>

        {/* Active Rooms Section */}
        {activeRooms.length > 0 && (
          <Card className="mt-12">
            <CardHeader>
              <CardTitle>Your Active Meetings</CardTitle>
              <CardDescription>Meetings you&apos;ve joined or created</CardDescription>
            </CardHeader>
            <CardContent className="divide-y">
              {activeRooms.map((room) => (
                <div key={room.id} className="py-4 first:pt-0 last:pb-0">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="font-semibold text-lg mb-2">Meeting {room.id.substring(0, 8)}</h3>
                      <div className="flex items-center gap-3">
                        <Badge variant="secondary">
                          {room.participants?.length || 0} participants
                        </Badge>
                        <Badge variant={room.status === 'active' ? 'default' : 'secondary'}>
                          {room.status}
                        </Badge>
                      </div>
                    </div>
                    <Link href={`/room/${room.id}`}>
                      <Button>
                        Join
                      </Button>
                    </Link>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        )}

        {isLoadingRooms && activeRooms.length === 0 && (
          <div className="mt-12 bg-white rounded-lg p-8 text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading your meetings...</p>
          </div>
        )}

        <div className="mt-12 bg-white rounded-lg p-6 border">
          <h2 className="text-xl font-semibold mb-4">How to Use</h2>
          <ul className="space-y-3 text-gray-700">
            <li className="flex gap-3">
              <span className="text-blue-600 font-semibold">1.</span>
              <span>Create a new meeting or join an existing one</span>
            </li>
            <li className="flex gap-3">
              <span className="text-blue-600 font-semibold">2.</span>
              <span>Allow access to your camera and microphone</span>
            </li>
            <li className="flex gap-3">
              <span className="text-blue-600 font-semibold">3.</span>
              <span>Share the meeting ID with participants</span>
            </li>
            <li className="flex gap-3">
              <span className="text-blue-600 font-semibold">4.</span>
              <span>Use controls to toggle camera, mic, and chat</span>
            </li>
          </ul>
        </div>
      </main>
    </div>
  );
}
