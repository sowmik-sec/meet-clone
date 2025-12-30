import { useState } from 'react';
import { useMediaStore } from '@/store/mediaStore';

export function useMediaDevices() {
  const { setLocalStream } = useMediaStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getMediaStream = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: {
          width: { ideal: 1280 },
          height: { ideal: 720 },
        },
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
        },
      });

      setLocalStream(stream);
      return stream;
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to access camera/microphone';
      setError(message);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const stopMediaStream = (stream: MediaStream) => {
    stream.getTracks().forEach((track) => track.stop());
    setLocalStream(null);
  };

  return {
    getMediaStream,
    stopMediaStream,
    isLoading,
    error,
  };
}
