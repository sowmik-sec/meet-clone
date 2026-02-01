'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useRealtimeKitClient, RealtimeKitProvider, useRealtimeKitMeeting } from '@cloudflare/realtimekit-react';
import {
  RtkMeeting
} from '@cloudflare/realtimekit-react-ui';
import { useAuthStore } from '@/store/authStore';
import { roomApi } from '@/lib/api/room';
import { callsApi } from '@/lib/api/calls';
import { AxiosError } from 'axios';
import { useBandwidthStats } from '@/hooks/useBandwidthStats';

function MeetingUI() {
  const { meeting } = useRealtimeKitMeeting();
  const params = useParams();
  const roomId = params.id as string;

  // Start collecting bandwidth stats
  useBandwidthStats(roomId);

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

  const { user, isAuthenticated, hasHydrated, initializeAuth } = useAuthStore();
  const [meeting, initMeeting] = useRealtimeKitClient();
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);

  // Initialize auth from cookies on mount
  useEffect(() => {
    initializeAuth();
  }, [initializeAuth]);

  // Wait for auth store to hydrate before checking authentication
  useEffect(() => {
    if (hasHydrated) {
      setIsCheckingAuth(false);
    }
  }, [hasHydrated]);

  useEffect(() => {
    // Don't do anything until auth store has hydrated
    if (isCheckingAuth) {
      return;
    }

    if (!isAuthenticated) {
      // Store the current room URL and redirect to login
      const currentPath = `/room/${roomId}`;
      router.push(`/login?redirect=${encodeURIComponent(currentPath)}`);
      return;
    }

    const initRoom = async () => {
      try {
        // Try to join the room
        try {
          await roomApi.joinRoom(roomId, {
            user_name: user?.name || 'Anonymous User',
            avatar: user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=' + (user?.email || 'default'),
          });
        } catch (joinError) {
          const axiosError = joinError as AxiosError<{ error?: string; message?: string }>;

          // If join fails with 404, room doesn't exist
          if (axiosError.response?.status === 404) {
            console.error('Room not found.');
            alert('This meeting does not exist. Please create a new meeting or use a valid meeting ID.');
            router.push('/dashboard');
            return;
          }

          // If join fails with 400, user might already be in the room or room validation failed
          if (axiosError.response?.status === 400) {
            const errorMsg = axiosError.response?.data?.error || axiosError.response?.data?.message;
            console.log('Join room validation error:', errorMsg);
            // Continue anyway to try getting the session - user might already be in the room
          } else {
            throw joinError;
          }
        }

        // Create/Get Cloudflare Session
        const session = await callsApi.createSession(roomId);
        const { token: cfToken } = await callsApi.generateToken(session.sessionId);

        // Initialize Cloudflare RealtimeKit
        initMeeting({ authToken: cfToken });
      } catch (error) {
        console.error('Error initializing room:', error);
        const axiosError = error as AxiosError;
        const errorMsg = axiosError.response?.status === 404
          ? 'Meeting not found. Please use a valid meeting ID.'
          : 'Failed to join the meeting. Please try again.';
        alert(errorMsg);
        router.push('/dashboard');
      }
    };

    initRoom();

    // Cleanup function when component unmounts
    return () => {
      if (isAuthenticated) {
        roomApi.leaveRoom(roomId).catch(err => {
          console.error('Failed to leave room on unmount:', err);
        });
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId, isAuthenticated, isCheckingAuth]);

  // Show loading while checking authentication
  if (isCheckingAuth) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-white mx-auto"></div>
          <h2 className="text-2xl font-bold text-white mt-4">Loading...</h2>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <RealtimeKitProvider value={meeting}>
      <MeetingUI />
    </RealtimeKitProvider>
  );
}
