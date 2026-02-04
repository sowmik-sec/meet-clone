'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useRealtimeKitClient, RealtimeKitProvider, useRealtimeKitMeeting } from '@cloudflare/realtimekit-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import {
  RtkMeeting
} from '@cloudflare/realtimekit-react-ui';
import { useAuthStore } from '@/store/authStore';
import { roomApi } from '@/lib/api/room';
import { callsApi } from '@/lib/api/calls';
import { AxiosError } from 'axios';
import { useBandwidthStats } from '@/hooks/useBandwidthStats';
import { WaitingParticipant } from '@/types/room';
import { useToast } from '@/components/ui/use-toast';

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
  const { toast } = useToast();
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);
  const [isWaiting, setIsWaiting] = useState(false);
  const [isCreator, setIsCreator] = useState(false);
  const [waitingParticipants, setWaitingParticipants] = useState<WaitingParticipant[]>([]);
  const [approvalPollInterval, setApprovalPollInterval] = useState<NodeJS.Timeout | null>(null);

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
        let roomData;
        try {
          roomData = await roomApi.joinRoom(roomId, {
            user_name: user?.name || 'Anonymous User',
            avatar: user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=' + (user?.email || 'default'),
          });
          setIsCreator(roomData.created_by === user?.id);
        } catch (joinError) {
          const axiosError = joinError as AxiosError<{ error?: string; message?: string }>;

          // If join fails with 404, room doesn't exist
          if (axiosError.response?.status === 404) {
            console.error('Room not found.');
            toast({
              title: "Meeting Not Found",
              description: "This meeting does not exist. Please create a new meeting.",
              variant: "destructive",
            });
            router.push('/dashboard');
            return;
          }

          // If join fails with 400, user might already be in the room or room validation failed
          if (axiosError.response?.status === 400) {
            const errorMsg = axiosError.response?.data?.error || axiosError.response?.data?.message;
            console.log('Join room validation error:', errorMsg);
          } else {
            throw joinError;
          }
        }

        // Create/Get Cloudflare Session
        const session = await callsApi.createSession(roomId);

        // Try to get token (may fail if not approved)
        try {
          const { token: cfToken } = await callsApi.generateToken(session.sessionId);
          // Initialize Cloudflare RealtimeKit
          initMeeting({ authToken: cfToken });
        } catch (tokenError) {
          const axiosError = tokenError as AxiosError<{ error?: string }>;

          // 403 means waiting for approval
          if (axiosError.response?.status === 403) {
            setIsWaiting(true);
            // Start polling for approval
            const pollInterval = setInterval(async () => {
              try {
                const { token: cfToken } = await callsApi.generateToken(session.sessionId);
                clearInterval(pollInterval);
                setApprovalPollInterval(null);
                setIsWaiting(false);
                initMeeting({ authToken: cfToken });
              } catch (e) {
                // Still getting 403, check if we were denied
                // Get room data to see if we're still in waiting list
                try {
                  const roomData = await roomApi.getRoom(roomId);
                  const isInWaiting = roomData.waiting_participants?.some(p => p.user_id === user?.id);
                  const isParticipant = roomData.participants?.some(p => p.user_id === user?.id && !p.left_at);

                  // If not in waiting list and not a participant, we were denied
                  if (!isInWaiting && !isParticipant) {
                    clearInterval(pollInterval);
                    setApprovalPollInterval(null);
                    toast({
                      title: "Access Denied",
                      description: "The host denied your request to join this meeting.",
                      variant: "destructive",
                    });
                    router.push('/dashboard');
                  }
                } catch (roomError) {
                  console.error('Failed to check room status:', roomError);
                }
              }
            }, 2000);

            setApprovalPollInterval(pollInterval);
          } else {
            throw tokenError;
          }
        }
      } catch (error) {
        console.error('Error initializing room:', error);
        const axiosError = error as AxiosError;
        const errorMsg = axiosError.response?.status === 404
          ? 'Meeting not found. Please use a valid meeting ID.'
          : 'Failed to join the meeting. Please try again.';
        toast({
          title: "Join Failed",
          description: errorMsg,
          variant: "destructive",
        });
        router.push('/dashboard');
      }
    };

    initRoom();

    // Poll for waiting participants if creator
    const waitingPollInterval = setInterval(async () => {
      try {
        const roomData = await roomApi.getRoom(roomId);
        if (roomData.created_by === user?.id) {
          setWaitingParticipants(roomData.waiting_participants || []);
        }
      } catch (err) {
        console.error('Failed to fetch waiting participants:', err);
      }
    }, 2000);

    // Cleanup function when component unmounts
    return () => {
      if (approvalPollInterval) {
        clearInterval(approvalPollInterval);
      }
      clearInterval(waitingPollInterval);
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

  const handleApprove = async (userId: string) => {
    try {
      await roomApi.approveParticipant(roomId, userId);
      setWaitingParticipants(prev => prev.filter(p => p.user_id !== userId));
    } catch (error) {
      console.error('Failed to approve participant:', error);
    }
  };

  const handleDeny = async (userId: string) => {
    try {
      await roomApi.denyParticipant(roomId, userId);
      setWaitingParticipants(prev => prev.filter(p => p.user_id !== userId));
    } catch (error) {
      console.error('Failed to deny participant:', error);
    }
  };

  // Show waiting screen if not approved
  if (isWaiting) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-white mx-auto mb-4"></div>
          <h2 className="text-2xl font-bold text-white">Waiting for host approval...</h2>
          <p className="text-gray-400 mt-2">The host will let you in soon</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <RealtimeKitProvider value={meeting}>
        <MeetingUI />
      </RealtimeKitProvider>

      {/* Waiting participants panel for creator */}
      {isCreator && waitingParticipants.length > 0 && (
        <Card className="fixed bottom-4 right-4 w-80 max-h-96 overflow-y-auto z-50 bg-gray-800 border-gray-700 shadow-xl">
          <CardHeader className="p-4 pb-2">
            <CardTitle className="text-white text-lg font-bold">Waiting Room ({waitingParticipants.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-4 pt-2 space-y-2">
            {waitingParticipants.map((participant) => (
              <div key={participant.user_id} className="flex items-center justify-between bg-gray-700 p-3 rounded-lg">
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={participant.avatar} alt={participant.name} />
                    <AvatarFallback>{participant.name.charAt(0)}</AvatarFallback>
                  </Avatar>
                  <span className="text-white text-sm font-medium">{participant.name}</span>
                </div>
                <div className="flex gap-2">
                  <Button
                    onClick={() => handleApprove(participant.user_id)}
                    size="sm"
                    className="h-8 px-2 bg-green-600 hover:bg-green-700 text-white"
                  >
                    Admit
                  </Button>
                  <Button
                    onClick={() => handleDeny(participant.user_id)}
                    size="sm"
                    variant="destructive"
                    className="h-8 px-2"
                  >
                    Deny
                  </Button>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </>
  );
}
