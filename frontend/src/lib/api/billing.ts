import api from './client';
import { UserBillingPeriod } from '@/types/billing';

export const billingApi = {
    getCurrentUsage: async (): Promise<UserBillingPeriod> => {
        const response = await api.get<UserBillingPeriod>('/billing/usage');
        return response.data;
    },

    getUsageByPeriod: async (period: string): Promise<UserBillingPeriod> => {
        const response = await api.get<UserBillingPeriod>(`/billing/usage/${period}`);
        return response.data;
    },

    getHistory: async (limit: number = 12): Promise<UserBillingPeriod[]> => {
        const response = await api.get<UserBillingPeriod[]>(`/billing/history?limit=${limit}`);
        return response.data;
    },

    endSession: async (roomId: string): Promise<void> => {
        await api.post('/billing/session/end', { roomId });
    },
};
