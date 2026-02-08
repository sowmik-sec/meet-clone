import api from './client';

export interface PublicProfile {
    id: string;
    name: string;
    avatar: string;
    bio: string;
}

export const userApi = {
    getPublicProfile: async (userId: string): Promise<PublicProfile> => {
        const res = await api.get<PublicProfile>(`/users/${userId}/profile`);
        return res.data;
    },

    updateProfile: async (data: { name: string; bio: string }) => {
        const res = await api.put<PublicProfile>('/auth/me', data);
        return res.data;
    }
};
