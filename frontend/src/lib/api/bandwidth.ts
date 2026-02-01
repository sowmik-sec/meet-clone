import api from './client';

export interface BandwidthReport {
    room_id: string;
    session_id: string;
    bytes_sent: number;
    bytes_received: number;
    packets_sent: number;
    packets_lost: number;
    duration: number;
}

export interface UserBandwidthStats {
    user_id: string;
    total_bytes_sent: number;
    total_bytes_received: number;
    total_duration: number;
    meeting_count: number;
}

export interface BandwidthRecord {
    id: string;
    user_id: string;
    room_id: string;
    session_id: string;
    bytes_sent: number;
    bytes_received: number;
    packets_sent: number;
    packets_lost: number;
    duration: number;
    created_at: string;
    updated_at: string;
}

export const bandwidthApi = {
    reportBandwidth: async (data: BandwidthReport): Promise<void> => {
        await api.post('/bandwidth/report', data);
    },

    getStats: async (): Promise<UserBandwidthStats> => {
        const response = await api.get<UserBandwidthStats>('/bandwidth/stats');
        return response.data;
    },

    getHistory: async (limit: number = 10): Promise<BandwidthRecord[]> => {
        const response = await api.get<BandwidthRecord[]>(`/bandwidth/history?limit=${limit}`);
        return response.data;
    },
};
