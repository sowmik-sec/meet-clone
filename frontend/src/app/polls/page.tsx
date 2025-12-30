'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useRealtimeKitClient, RealtimeKitProvider, useRealtimeKitMeeting } from '@cloudflare/realtimekit-react';
import { 
  RtkUiProvider,
  RtkPolls,
  RtkParticipantsAudio,
  RtkNotifications
} from '@cloudflare/realtimekit-react-ui';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';

function PollsView() {
  const { meeting } = useRealtimeKitMeeting();
  const router = useRouter();
  
  if (!meeting) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center text-white">
          <p>Loading polls...</p>
        </div>
      </div>
    );
  }

  return (
    <RtkUiProvider meeting={meeting} style={{ height: '100vh', backgroundColor: '#1a1a1a' }}>
      <RtkParticipantsAudio />
      <RtkNotifications />
      
      <div className="h-full flex flex-col">
        <div className="bg-gray-800 border-b border-gray-700 p-4 flex items-center justify-between">
          <h1 className="text-white text-xl font-bold">Meeting Polls</h1>
          <Button
            variant="ghost"
            onClick={() => router.back()}
          >
            Back to Meeting
          </Button>
        </div>
        
        <div className="flex-1 overflow-auto p-6">
          <RtkPolls style={{ maxWidth: '800px', margin: '0 auto' }} />
        </div>
      </div>
    </RtkUiProvider>
  );
}

export default function PollsPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const [meeting] = useRealtimeKitClient();

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <RealtimeKitProvider value={meeting}>
      <PollsView />
    </RealtimeKitProvider>
  );
}
