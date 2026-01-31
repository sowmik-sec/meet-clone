'use client';

import { useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useRealtimeKitClient, RealtimeKitProvider, useRealtimeKitMeeting } from '@cloudflare/realtimekit-react';
import {
  RtkMeeting
} from '@cloudflare/realtimekit-react-ui';
import { useAuthStore } from '@/store/authStore';
import { roomApi } from '@/lib/api/room';
import { callsApi } from '@/lib/api/calls';

function MeetingUI() {
  const { meeting } = useRealtimeKitMeeting();

  if (!meeting) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-white mx-auto"></div>
          <h2 className="text-2xl font-bold text-white mt-4">Loading meeting...</h2>
        </div>
      </div>
    );
  }

  return <RtkMeeting meeting={meeting} />;
}

export default function RoomPage() {
  const router = useRouter();
  const params = useParams();
  const roomId = params.id as string;

  const { user, isAuthenticated } = useAuthStore();
  const [meeting, initMeeting] = useRealtimeKitClient();

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
      return;
    }

    const initRoom = async () => {
      try {
        // Join room through backend
        await roomApi.joinRoom(roomId, {
          user_name: user?.name || 'Unknown',
          avatar: user?.avatar || '',
        });

        // Create/Get Cloudflare Session
        const session = await callsApi.createSession(roomId);
        const { token: cfToken } = await callsApi.generateToken(session.sessionId);

        // Initialize Cloudflare RealtimeKit
        initMeeting({ authToken: cfToken });
      } catch (error: unknown) {
        console.error('Error initializing room:', error);
        router.push('/dashboard');
      }
    };

    initRoom();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId, isAuthenticated]);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <RealtimeKitProvider value={meeting}>
      <MeetingUI />
    </RealtimeKitProvider>
  );
}
