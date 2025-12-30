import { create } from 'zustand';
import { Room, Participant } from '@/types/room';

interface RoomState {
  room: Room | null;
  participants: Participant[];
  isInRoom: boolean;
  setRoom: (room: Room) => void;
  setParticipants: (participants: Participant[]) => void;
  addParticipant: (participant: Participant) => void;
  removeParticipant: (userId: string) => void;
  clearRoom: () => void;
}

export const useRoomStore = create<RoomState>((set) => ({
  room: null,
  participants: [],
  isInRoom: false,
  setRoom: (room) => set({ room, isInRoom: true }),
  setParticipants: (participants) => set({ participants }),
  addParticipant: (participant) =>
    set((state) => ({
      participants: [...state.participants, participant],
    })),
  removeParticipant: (userId) =>
    set((state) => ({
      participants: state.participants.filter((p) => p.user_id !== userId),
    })),
  clearRoom: () => set({ room: null, participants: [], isInRoom: false }),
}));
