export interface Appointment {
    id: string;
    host_id: string;
    guest_id: string;
    room_id?: string;
    title: string;
    description?: string;
    start_time: string;
    end_time: string;
    timezone: string;
    status: 'pending' | 'confirmed' | 'cancelled' | 'completed';
    meeting_type: 'meeting' | 'webinar';
    created_at: string;
    updated_at: string;
}

export interface CreateAppointmentRequest {
    guest_id?: string;
    title: string;
    description?: string;
    start_time: string;
    end_time: string;
    timezone: string;
    meeting_type: 'meeting' | 'webinar';
}

export interface UpdateAppointmentRequest {
    title?: string;
    description?: string;
    start_time?: string;
    end_time?: string;
    status?: Appointment['status'];
}

export interface AppointmentFilter {
    start_time_after?: string;
    start_time_before?: string;
    status?: Appointment['status'];
}
