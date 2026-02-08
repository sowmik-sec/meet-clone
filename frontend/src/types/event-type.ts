export interface EventType {
    id: string;
    user_id: string;
    title: string;
    slug: string;
    description: string;
    duration: number;
    buffer_before: number;
    buffer_after: number;
    color: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    min_cancel_notice: number;
    min_reschedule_notice: number;
    allow_guest_cancel: boolean;
    max_attendees: number;
}

export interface CreateEventTypeRequest {
    title: string;
    slug: string;
    description: string;
    duration: number;
    buffer_before: number;
    buffer_after: number;
    color: string;
    min_cancel_notice: number;
    min_reschedule_notice: number;
    allow_guest_cancel: boolean;
    max_attendees: number;
}

export interface UpdateEventTypeRequest {
    title?: string;
    slug?: string;
    description?: string;
    duration?: number;
    buffer_before?: number;
    buffer_after?: number;
    color?: string;
    is_active?: boolean;
    min_cancel_notice?: number;
    min_reschedule_notice?: number;
    allow_guest_cancel?: boolean;
    max_attendees?: number;
}
