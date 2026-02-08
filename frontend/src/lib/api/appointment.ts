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

    createPublicBooking: async (userId: string, data: { guest_name: string; guest_email: string; start_time: string; timezone: string; event_type_id?: string; answers?: Record<string, string> }): Promise<Appointment> => {
        const response = await api.post<Appointment>(`/users/${userId}/bookings`, data);
        return response.data;
    },
    getBookedSlots: async (userId: string, date: string, eventTypeId?: string): Promise<{ busy_slots: string[][], partial_slots: Record<string, number> }> => {
        let url = `/users/${userId}/booked-slots?date=${date}`;
        if (eventTypeId) {
            url += `&event_type_id=${eventTypeId}`;
        }
        const response = await api.get<{ busy_slots: string[][], partial_slots: Record<string, number> }>(url);
        return response.data;
    },

    getAppointmentByRescheduleToken: async (token: string): Promise<Appointment> => {
        const response = await api.get<Appointment>(`/appointments/reschedule/${token}`);
        return response.data;
    },

    rescheduleAppointment: async (token: string, newStartTime: string): Promise<Appointment> => {
        const response = await api.post<Appointment>(`/appointments/reschedule/${token}`, { new_start_time: newStartTime });
        return response.data;
    },

    cancelAppointmentByToken: async (token: string): Promise<void> => {
        await api.delete(`/appointments/reschedule/${token}`);
    },

    getAppointment: async (id: string): Promise<Appointment> => {
        const response = await api.get<Appointment>(`/appointments/${id}`);
        return response.data;
    },
};
