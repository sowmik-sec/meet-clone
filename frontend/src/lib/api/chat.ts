import api from './client';
import { Message } from '@/types/chat';

export const chatApi = {
  getMessages: async (roomId: string, limit = 50, offset = 0): Promise<Message[]> => {
    const response = await api.get<Message[]>(`/rooms/${roomId}/messages`, {
      params: { limit, offset },
    });
    return response.data;
  },
};
