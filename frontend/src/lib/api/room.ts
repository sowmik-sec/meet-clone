import api from './client';
import { Room, JoinRoomRequest, Participant } from '@/types/room';

export const roomApi = {
  createRoom: async (): Promise<Room> => {
    const response = await api.post<Room>('/rooms');
    return response.data;
  },

  getRoom: async (roomId: string): Promise<Room> => {
    const response = await api.get<Room>(`/rooms/${roomId}`);
    return response.data;
  },

  joinRoom: async (roomId: string, data: JoinRoomRequest): Promise<Room> => {
    const response = await api.post<Room>(`/rooms/${roomId}/join`, data);
    return response.data;
  },

  leaveRoom: async (roomId: string): Promise<Room> => {
    const response = await api.post<Room>(`/rooms/${roomId}/leave`);
    return response.data;
  },

  endRoom: async (roomId: string): Promise<void> => {
    await api.delete(`/rooms/${roomId}`);
  },

  getParticipants: async (roomId: string): Promise<Participant[]> => {
    const response = await api.get<Participant[]>(`/rooms/${roomId}/participants`);
    return response.data;
  },

  getUserRooms: async (): Promise<Room[]> => {
    const response = await api.get<Room[]>('/rooms/my-rooms');
    return response.data;
  },
};
