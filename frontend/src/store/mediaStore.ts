import { create } from 'zustand';

interface MediaState {
  isCameraOn: boolean;
  isMicOn: boolean;
  localStream: MediaStream | null;
  toggleCamera: () => void;
  toggleMic: () => void;
  setLocalStream: (stream: MediaStream | null) => void;
}

export const useMediaStore = create<MediaState>((set) => ({
  isCameraOn: true,
  isMicOn: true,
  localStream: null,
  toggleCamera: () =>
    set((state) => {
      if (state.localStream) {
        state.localStream.getVideoTracks().forEach((track) => {
          track.enabled = !state.isCameraOn;
        });
      }
      return { isCameraOn: !state.isCameraOn };
    }),
  toggleMic: () =>
    set((state) => {
      if (state.localStream) {
        state.localStream.getAudioTracks().forEach((track) => {
          track.enabled = !state.isMicOn;
        });
      }
      return { isMicOn: !state.isMicOn };
    }),
  setLocalStream: (stream) => set({ localStream: stream }),
}));
