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
}

export interface CreateEventTypeRequest {
    title: string;
    slug: string;
    description: string;
    duration: number;
    buffer_before: number;
    buffer_after: number;
    color: string;
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
}
