'use client';

import { useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';

export default function WaitingRoomPage() {
  const router = useRouter();
  const params = useParams();
  const roomId = params.id as string;
  const { user, isAuthenticated } = useAuthStore();
  const [isJoining, setIsJoining] = useState(false);

  const handleJoinMeeting = async () => {
    setIsJoining(true);
    // Navigate to actual room
    router.push(`/room/${roomId}`);
  };

  if (!isAuthenticated) {
    router.push('/login');
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-blue-900 to-gray-900 flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <CardHeader className="text-center">
          <div className="w-20 h-20 bg-blue-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </div>
          <CardTitle className="text-3xl">Ready to join?</CardTitle>
          <CardDescription>
            Meeting ID: <Badge variant="secondary">{roomId}</Badge>
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center space-x-3">
                <Avatar className="h-12 w-12">
                  <AvatarFallback className="bg-blue-500 text-white font-semibold">
                    {user?.name?.charAt(0).toUpperCase() || 'U'}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium">{user?.name || 'User'}</p>
                  <p className="text-sm text-muted-foreground">{user?.email}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Meeting Features</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2 text-sm">
                <li className="flex items-center gap-2">
                  <Badge variant="outline" className="w-fit h-fit p-0.5">✓</Badge>
                  HD Video & Audio
                </li>
                <li className="flex items-center gap-2">
                  <Badge variant="outline" className="w-fit h-fit p-0.5">✓</Badge>
                  Real-time Chat
                </li>
                <li className="flex items-center gap-2">
                  <Badge variant="outline" className="w-fit h-fit p-0.5">✓</Badge>
                  Screen Sharing
                </li>
                <li className="flex items-center gap-2">
                  <Badge variant="outline" className="w-fit h-fit p-0.5">✓</Badge>
                  Recording
                </li>
              </ul>
            </CardContent>
          </Card>

          <Button
            onClick={handleJoinMeeting}
            disabled={isJoining}
            className="w-full"
          >
            {isJoining ? 'Joining...' : 'Join Meeting Now'}
          </Button>

          <Button
            onClick={() => router.push('/dashboard')}
            variant="ghost"
            className="w-full"
          >
            Back to Dashboard
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
