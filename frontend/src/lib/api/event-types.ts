import { api } from './client';
import { EventType, CreateEventTypeRequest, UpdateEventTypeRequest } from '@/types/event-type';

export const eventTypesApi = {
    list: async () => {
        const response = await api.get<EventType[]>('/event-types');
        return response.data;
    },

    create: async (data: CreateEventTypeRequest) => {
        const response = await api.post<EventType>('/event-types', data);
        return response.data;
    },

    update: async (id: string, data: UpdateEventTypeRequest) => {
        const response = await api.put<EventType>(`/event-types/${id}`, data);
        return response.data;
    },

    delete: async (id: string) => {
        await api.delete(`/event-types/${id}`);
    },

    listPublic: async (userId: string) => {
        const response = await api.get<EventType[]>(`/users/${userId}/event-types`);
        return response.data;
    },
};
