"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { CreateEventTypeRequest, EventType, UpdateEventTypeRequest } from "@/types/event-type";
import { eventTypesApi } from "@/lib/api/event-types";
import { toast } from "sonner"; // Assuming sonner or useToast

interface EventTypeFormProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    eventType?: EventType; // If provided, edit mode
    onSuccess: () => void;
}

export function EventTypeForm({ open, onOpenChange, eventType, onSuccess }: EventTypeFormProps) {
    const [loading, setLoading] = useState(false);

    const { register, handleSubmit, formState: { errors }, reset, setValue, watch } = useForm<CreateEventTypeRequest>({
        defaultValues: eventType ? {
            title: eventType.title,
            slug: eventType.slug,
            description: eventType.description,
            duration: eventType.duration,
            buffer_before: eventType.buffer_before,
            buffer_after: eventType.buffer_after,
            color: eventType.color,
        } : {
            duration: 30,
            buffer_before: 0,
            buffer_after: 0,
            color: "#000000",
        }
    });

    const onSubmit = async (data: CreateEventTypeRequest) => {
        setLoading(true);
        try {
            if (eventType) {
                await eventTypesApi.update(eventType.id, data);
                toast.success("Event type updated successfully");
            } else {
                await eventTypesApi.create(data);
                toast.success("Event type created successfully");
            }
            onSuccess();
            onOpenChange(false);
            reset();
        } catch (error) {
            toast.error("Something went wrong");
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>{eventType ? "Edit Event Type" : "Create Event Type"}</DialogTitle>
                    <DialogDescription>
                        {eventType ? "Update existing event type details." : "Add a new meeting type for people to book."}
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                    <div className="grid gap-2">
                        <Label htmlFor="title">Title</Label>
                        <Input id="title" {...register("title", { required: true })} placeholder="e.g. Quick Chat" />
                        {errors.title && <span className="text-sm text-red-500">Title is required</span>}
                    </div>

                    <div className="grid gap-2">
                        <Label htmlFor="slug">URL Slug</Label>
                        <Input id="slug" {...register("slug", { required: true })} placeholder="e.g. quick-chat" />
                        {errors.slug && <span className="text-sm text-red-500">Slug is required</span>}
                    </div>

                    <div className="grid gap-2">
                        <Label htmlFor="duration">Duration (minutes)</Label>
                        <Select
                            onValueChange={(val: string) => setValue("duration", parseInt(val))}
                            defaultValue={eventType?.duration?.toString() || "30"}
                        >
                            <SelectTrigger>
                                <SelectValue placeholder="Select duration" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="15">15 min</SelectItem>
                                <SelectItem value="30">30 min</SelectItem>
                                <SelectItem value="45">45 min</SelectItem>
                                <SelectItem value="60">60 min</SelectItem>
                                <SelectItem value="90">90 min</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="grid gap-2">
                            <Label htmlFor="buffer_before">Buffer Before</Label>
                            <Select
                                onValueChange={(val: string) => setValue("buffer_before", parseInt(val))}
                                defaultValue={eventType?.buffer_before?.toString() || "0"}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="None" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="0">None</SelectItem>
                                    <SelectItem value="5">5 min</SelectItem>
                                    <SelectItem value="10">10 min</SelectItem>
                                    <SelectItem value="15">15 min</SelectItem>
                                    <SelectItem value="30">30 min</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="buffer_after">Buffer After</Label>
                            <Select
                                onValueChange={(val: string) => setValue("buffer_after", parseInt(val))}
                                defaultValue={eventType?.buffer_after?.toString() || "0"}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="None" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="0">None</SelectItem>
                                    <SelectItem value="5">5 min</SelectItem>
                                    <SelectItem value="10">10 min</SelectItem>
                                    <SelectItem value="15">15 min</SelectItem>
                                    <SelectItem value="30">30 min</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    <div className="grid gap-2">
                        <Label htmlFor="description">Description</Label>
                        <Textarea id="description" {...register("description")} placeholder="Add details about this event type..." />
                    </div>

                    <DialogFooter>
                        <Button type="submit" disabled={loading}>{loading ? "Saving..." : "Save Event Type"}</Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
