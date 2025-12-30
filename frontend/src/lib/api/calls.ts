import { api } from './client';

export const callsApi = {
  createSession: async (roomId: string) => {
    const response = await api.post('/calls/sessions', {
      roomId,
    });
    return response.data;
  },

  generateToken: async (sessionId: string) => {
    const response = await api.post('/calls/sessions/token', {
      sessionId,
    });
    return response.data;
  },
};
