import api from './client';
import { Appointment, CreateAppointmentRequest, UpdateAppointmentRequest, AppointmentFilter } from '@/types/appointment';

export const appointmentApi = {
    createAppointment: async (data: CreateAppointmentRequest): Promise<Appointment> => {
        const response = await api.post<Appointment>('/appointments', data);
        return response.data;
    },

    getAppointments: async (filter?: AppointmentFilter): Promise<Appointment[]> => {
        const params = new URLSearchParams();
        if (filter?.status) params.append('status', filter.status);
        if (filter?.start_time_after) params.append('start_time_after', filter.start_time_after);
        if (filter?.start_time_before) params.append('start_time_before', filter.start_time_before);

        const response = await api.get<Appointment[]>(`/appointments?${params.toString()}`);
        return response.data;
    },

    confirmAppointment: async (id: string): Promise<void> => {
        await api.post(`/appointments/${id}/confirm`);
    },

    cancelAppointment: async (id: string): Promise<void> => {
        await api.delete(`/appointments/${id}`);
    },

    startAppointment: async (id: string): Promise<{ room_id: string }> => {
        const response = await api.post<{ room_id: string }>(`/appointments/${id}/start`);
        return response.data;
    },
};
