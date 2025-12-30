'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useRealtimeKitClient, RealtimeKitProvider, useRealtimeKitMeeting } from '@cloudflare/realtimekit-react';
import {
  RtkUiProvider,
  RtkHeader,
  RtkGrid,
  RtkSidebar,
  RtkControlbar,
  RtkParticipantsAudio,
  RtkNotifications,
  RtkDialogManager,
  RtkSetupScreen,
  RtkEndedScreen,
  RtkWaitingScreen,
  RtkChat,
  RtkParticipantTile,
  RtkParticipantCount,
  RtkMeetingTitle,
  RtkRecordingIndicator,
  RtkClock,
  RtkLeaveMeeting,
  RtkSpotlightGrid,
  RtkSimpleGrid,
  RtkMixedGrid,
  RtkSettingsAudio,
  RtkSettingsVideo,
  RtkNameTag,
  RtkSpinner,
  RtkTooltip,
  RtkAudioVisualizer
} from '@cloudflare/realtimekit-react-ui';
import { useAuthStore } from '@/store/authStore';
import { roomApi } from '@/lib/api/room';
import { callsApi } from '@/lib/api/calls';
import { Button } from '@/components/ui/button';

function MeetingUI() {
  const { meeting } = useRealtimeKitMeeting();
  const [meetingState, setMeetingState] = useState<string>('idle');
  const [showChat, setShowChat] = useState(false);
  const [gridType, setGridType] = useState<'default' | 'spotlight' | 'simple' | 'mixed'>('default');

  if (!meeting) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <RtkSpinner />
          <h2 className="text-2xl font-bold text-white mt-4">Loading meeting...</h2>
        </div>
      </div>
    );
  }

  return (
    <RtkUiProvider
      meeting={meeting}
      showSetupScreen={true}
      onRtkStatesUpdate={(event: { detail?: { meeting?: string; activeSidebar?: boolean } }) => {
        if (event.detail?.meeting) {
          setMeetingState(event.detail.meeting);
        }
        if (event.detail?.activeSidebar !== undefined) {
          setShowChat(event.detail.activeSidebar);
        }
      }}
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100vh',
        backgroundColor: '#1a1a1a'
      }}
    >
      {/* Essential components - always rendered */}
      <RtkParticipantsAudio />
      <RtkNotifications />
      <RtkDialogManager />

      {/* Meeting states */}
      {meetingState === 'setup' && (
        <RtkSetupScreen />
      )}

      {meetingState === 'waiting' && (
        <RtkWaitingScreen />
      )}

      {meetingState === 'joined' && (
        <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
          {/* Custom Header with Cloudflare components */}
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            padding: '1rem',
            backgroundColor: '#2d2d2d',
            borderBottom: '1px solid #444'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
              <RtkMeetingTitle />
              <RtkRecordingIndicator />
              <RtkParticipantCount />
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
              <RtkClock />
              <RtkTooltip content="Leave Meeting">
                <RtkLeaveMeeting />
              </RtkTooltip>
            </div>
          </div>

          {/* Main Content Area */}
          <div style={{ display: 'flex', flex: 1, position: 'relative', overflow: 'hidden' }}>
            {/* Video Grid with multiple layouts */}
            <div style={{ flex: 1, position: 'relative' }}>
              {gridType === 'default' && <RtkGrid />}
              {gridType === 'spotlight' && <RtkSpotlightGrid />}
              {gridType === 'simple' && <RtkSimpleGrid />}
              {gridType === 'mixed' && <RtkMixedGrid />}

              {/* Grid Type Selector */}
              <div style={{
                position: 'absolute',
                top: '1rem',
                right: showChat ? '370px' : '1rem',
                display: 'flex',
                gap: '0.5rem',
                backgroundColor: 'rgba(0,0,0,0.6)',
                padding: '0.5rem',
                borderRadius: '0.5rem'
              }}>
                <Button
                  variant={gridType === 'default' ? 'default' : 'secondary'}
                  size="sm"
                  onClick={() => setGridType('default')}
                >
                  Grid
                </Button>
                <Button
                  variant={gridType === 'spotlight' ? 'default' : 'secondary'}
                  size="sm"
                  onClick={() => setGridType('spotlight')}
                >
                  Spotlight
                </Button>
                <Button
                  variant={gridType === 'simple' ? 'default' : 'secondary'}
                  size="sm"
                  onClick={() => setGridType('simple')}
                >
                  Simple
                </Button>
                <Button
                  variant={gridType === 'mixed' ? 'default' : 'secondary'}
                  size="sm"
                  onClick={() => setGridType('mixed')}
                >
                  Mixed
                </Button>
              </div>
            </div>

            {/* Sidebar with Chat */}
            <RtkSidebar style={{
              position: 'absolute',
              right: 0,
              top: 0,
              bottom: 0,
              width: '350px'
            }} />
          </div>

          {/* Control Bar */}
          <RtkControlbar style={{
            display: 'flex',
            justifyContent: 'center',
            padding: '1rem',
            backgroundColor: '#2d2d2d'
          }} />
        </div>
      )}

      {meetingState === 'ended' && (
        <RtkEndedScreen />
      )}

      {meetingState === 'idle' && (
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <RtkSpinner />
            <p className="text-white text-lg mt-4">Initializing meeting...</p>
            <RtkAudioVisualizer style={{ marginTop: '2rem' }} />
          </div>
        </div>
      )}
    </RtkUiProvider>
  );
}

export default function RoomPage() {
  const router = useRouter();
  const params = useParams();
  const roomId = params.id as string;

  const { user, token, isAuthenticated } = useAuthStore();
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
