"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Plus, Clock, MoreVertical, Edit, Trash2 } from "lucide-react";
import { EventType } from "@/types/event-type";
import { eventTypesApi } from "@/lib/api/event-types";
import { EventTypeForm } from "@/components/settings/event-type-form";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";
import { useAuth } from "@/hooks/useAuth";

export default function EventTypesPage() {
    const { user } = useAuth();
    const [eventTypes, setEventTypes] = useState<EventType[]>([]);
    const [loading, setLoading] = useState(true);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [editingType, setEditingType] = useState<EventType | undefined>(undefined);

    const fetchEventTypes = async () => {
        try {
            const data = await eventTypesApi.list();
            setEventTypes(data);
        } catch (error) {
            console.error(error);
            toast.error("Failed to load event types");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchEventTypes();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this event type?")) return;
        try {
            await eventTypesApi.delete(id);
            toast.success("Event type deleted");
            fetchEventTypes();
        } catch (error) {
            toast.error("Failed to delete event type");
        }
    };

    return (
        <div className="container mx-auto py-10 px-4 max-w-5xl">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Event Types</h1>
                    <p className="text-muted-foreground mt-1">Create and manage your meeting types.</p>
                </div>
                <Button onClick={() => { setEditingType(undefined); setIsCreateOpen(true); }}>
                    <Plus className="mr-2 h-4 w-4" /> Create New
                </Button>
            </div>

            {loading ? (
                <div>Loading...</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {eventTypes.length === 0 && (
                        <div className="col-span-full text-center py-10 border rounded-lg border-dashed">
                            <p className="text-muted-foreground">No event types found. Create one to get started.</p>
                        </div>
                    )}
                    {eventTypes.map((et) => (
                        <div key={et.id} className="border rounded-lg p-4 bg-card text-card-foreground shadow-sm hover:shadow-md transition-shadow">
                            <div className="flex justify-between items-start">
                                <div>
                                    <h3 className="font-semibold text-lg">{et.title}</h3>
                                    <p className="text-sm text-muted-foreground flex items-center mt-1">
                                        <Clock className="w-3 h-3 mr-1" /> {et.duration} mins
                                        {et.is_active ? null : <span className="ml-2 px-1.5 py-0.5 bg-yellow-100 text-yellow-800 text-xs rounded-full">Hidden</span>}
                                    </p>
                                    <p className="text-sm text-muted-foreground mt-2 line-clamp-2">{et.description}</p>
                                </div>
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" size="icon" className="h-8 w-8">
                                            <MoreVertical className="h-4 w-4" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuItem onClick={() => { setEditingType(et); setIsCreateOpen(true); }}>
                                            <Edit className="w-4 h-4 mr-2" /> Edit
                                        </DropdownMenuItem>
                                        <DropdownMenuItem className="text-red-600 focus:text-red-600" onClick={() => handleDelete(et.id)}>
                                            <Trash2 className="w-4 h-4 mr-2" /> Delete
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </div>
                            <div className="mt-4 pt-4 border-t flex justify-between items-center text-sm">
                                <span className="text-muted-foreground">/{et.slug}</span>
                                <Button variant="outline" size="sm" onClick={() => {
                                    navigator.clipboard.writeText(`${window.location.origin}/b/${user?.id}/${et.slug}`);
                                    toast.success("Link copied");
                                }}>Copy Link</Button>
                                <Button variant="ghost" size="sm" onClick={() => {
                                    window.open(`/b/${user?.id}/${et.slug}`, '_blank');
                                }}>View</Button>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            <EventTypeForm
                open={isCreateOpen}
                onOpenChange={(open) => {
                    setIsCreateOpen(open);
                    if (!open) setEditingType(undefined);
                }}
                eventType={editingType}
                onSuccess={fetchEventTypes}
            />
        </div>
    );
}
