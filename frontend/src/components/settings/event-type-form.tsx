"use client";

import { useState, useEffect } from "react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CreateEventTypeRequest, EventType } from "@/types/event-type";
import { eventTypesApi } from "@/lib/api/event-types";
import { toast } from "sonner";
import { QuestionEditor } from "@/components/event-type/question-editor";

interface EventTypeFormProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    eventType?: EventType;
    onSuccess: () => void;
}

export function EventTypeForm({ open, onOpenChange, eventType, onSuccess }: EventTypeFormProps) {
    const [loading, setLoading] = useState(false);

    const { register, handleSubmit, formState: { errors }, reset, setValue, watch } = useForm<CreateEventTypeRequest>({
        defaultValues: {
            duration: 30,
            buffer_before: 0,
            buffer_after: 0,
            color: "#000000",
            min_cancel_notice: 0,
            min_reschedule_notice: 0,
            allow_guest_cancel: true,
            max_attendees: 1,
            questions: [],
        }
    });

    useEffect(() => {
        if (open) {
            reset(eventType ? {
                title: eventType.title,
                slug: eventType.slug,
                description: eventType.description,
                duration: eventType.duration,
                buffer_before: eventType.buffer_before,
                buffer_after: eventType.buffer_after,
                color: eventType.color,
                min_cancel_notice: eventType.min_cancel_notice,
                min_reschedule_notice: eventType.min_reschedule_notice,
                allow_guest_cancel: eventType.allow_guest_cancel,
                max_attendees: eventType.max_attendees,
                questions: eventType.questions || [],
            } : {
                duration: 30,
                buffer_before: 0,
                buffer_after: 0,
                color: "#000000",
                min_cancel_notice: 0,
                min_reschedule_notice: 0,
                allow_guest_cancel: true,
                max_attendees: 1,
                questions: [],
            });
        }
    }, [eventType, open, reset]);

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
            <DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-hidden flex flex-col">
                <DialogHeader>
                    <DialogTitle>{eventType ? "Edit Event Type" : "Create Event Type"}</DialogTitle>
                    <DialogDescription>
                        {eventType ? "Update details for this event type." : "Create a new meeting type for people to book."}
                    </DialogDescription>
                </DialogHeader>

                <form onSubmit={handleSubmit(onSubmit)} className="flex-1 overflow-hidden flex flex-col">
                    <Tabs defaultValue="general" className="flex-1 overflow-hidden flex flex-col">
                        <TabsList className="grid w-full grid-cols-3">
                            <TabsTrigger value="general">General</TabsTrigger>
                            <TabsTrigger value="scheduling">Scheduling</TabsTrigger>
                            <TabsTrigger value="questions">Questions</TabsTrigger>
                        </TabsList>

                        <div className="flex-1 overflow-y-auto py-4 px-1">
                            {/* General Tab */}
                            <TabsContent value="general" className="space-y-4 mt-0">
                                <div className="grid gap-2">
                                    <Label htmlFor="title">Title</Label>
                                    <Input id="title" {...register("title", { required: true })} placeholder="e.g. Quick Chat" />
                                    {errors.title && <span className="text-sm text-red-500">Title is required</span>}
                                </div>

                                <div className="grid gap-2">
                                    <Label htmlFor="slug">URL Slug</Label>
                                    <div className="flex items-center">
                                        <span className="text-sm text-gray-500 bg-gray-50 border border-r-0 rounded-l-md px-3 py-2 h-10">/</span>
                                        <Input id="slug" {...register("slug", { required: true })} className="rounded-l-none" placeholder="quick-chat" />
                                    </div>
                                    {errors.slug && <span className="text-sm text-red-500">Slug is required</span>}
                                </div>

                                <div className="grid gap-2">
                                    <Label htmlFor="description">Description</Label>
                                    <Textarea
                                        id="description"
                                        {...register("description")}
                                        placeholder="Add details about this event type..."
                                        className="min-h-[100px]"
                                    />
                                </div>
                            </TabsContent>

                            {/* Scheduling Tab */}
                            <TabsContent value="scheduling" className="space-y-6 mt-0">
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="grid gap-2">
                                        <Label htmlFor="duration">Duration</Label>
                                        <Select
                                            onValueChange={(val) => setValue("duration", parseInt(val))}
                                            defaultValue={eventType?.duration?.toString() || "30"}
                                        >
                                            <SelectTrigger><SelectValue placeholder="Select duration" /></SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="15">15 min</SelectItem>
                                                <SelectItem value="30">30 min</SelectItem>
                                                <SelectItem value="45">45 min</SelectItem>
                                                <SelectItem value="60">60 min</SelectItem>
                                                <SelectItem value="90">90 min</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    <div className="grid gap-2">
                                        <Label htmlFor="max_attendees">Max Attendees</Label>
                                        <Select
                                            onValueChange={(val) => setValue("max_attendees", parseInt(val))}
                                            defaultValue={eventType?.max_attendees?.toString() || "1"}
                                        >
                                            <SelectTrigger><SelectValue placeholder="1 (One-on-One)" /></SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="1">1 (One-on-One)</SelectItem>
                                                <SelectItem value="2">2 Attendees</SelectItem>
                                                <SelectItem value="5">5 Attendees</SelectItem>
                                                <SelectItem value="10">10 Attendees</SelectItem>
                                                <SelectItem value="20">20 Attendees</SelectItem>
                                                <SelectItem value="50">50 Attendees</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </div>

                                <div className="border-t pt-4">
                                    <h4 className="text-sm font-medium mb-3">Buffers</h4>
                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="grid gap-2">
                                            <Label htmlFor="buffer_before">Before Event</Label>
                                            <Select
                                                onValueChange={(val) => setValue("buffer_before", parseInt(val))}
                                                defaultValue={eventType?.buffer_before?.toString() || "0"}
                                            >
                                                <SelectTrigger><SelectValue placeholder="None" /></SelectTrigger>
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
                                            <Label htmlFor="buffer_after">After Event</Label>
                                            <Select
                                                onValueChange={(val) => setValue("buffer_after", parseInt(val))}
                                                defaultValue={eventType?.buffer_after?.toString() || "0"}
                                            >
                                                <SelectTrigger><SelectValue placeholder="None" /></SelectTrigger>
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
                                </div>

                                <div className="border-t pt-4">
                                    <h4 className="text-sm font-medium mb-3">Cancellation Policy</h4>
                                    <div className="grid grid-cols-2 gap-4 mb-4">
                                        <div className="grid gap-2">
                                            <Label>Min Cancel Notice</Label>
                                            <Select
                                                onValueChange={(val) => setValue("min_cancel_notice", parseInt(val))}
                                                defaultValue={eventType?.min_cancel_notice?.toString() || "0"}
                                            >
                                                <SelectTrigger><SelectValue placeholder="Anytime" /></SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="0">Anytime</SelectItem>
                                                    <SelectItem value="1">1 hour</SelectItem>
                                                    <SelectItem value="2">2 hours</SelectItem>
                                                    <SelectItem value="4">4 hours</SelectItem>
                                                    <SelectItem value="24">24 hours</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                        <div className="grid gap-2">
                                            <Label>Min Reschedule Notice</Label>
                                            <Select
                                                onValueChange={(val) => setValue("min_reschedule_notice", parseInt(val))}
                                                defaultValue={eventType?.min_reschedule_notice?.toString() || "0"}
                                            >
                                                <SelectTrigger><SelectValue placeholder="Anytime" /></SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="0">Anytime</SelectItem>
                                                    <SelectItem value="1">1 hour</SelectItem>
                                                    <SelectItem value="2">2 hours</SelectItem>
                                                    <SelectItem value="4">4 hours</SelectItem>
                                                    <SelectItem value="24">24 hours</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        <input
                                            type="checkbox"
                                            id="allow_guest_cancel"
                                            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                                            {...register("allow_guest_cancel")}
                                        />
                                        <Label htmlFor="allow_guest_cancel" className="font-normal cursor-pointer">
                                            Allow guests to cancel their booking
                                        </Label>
                                    </div>
                                </div>
                            </TabsContent>

                            {/* Questions Tab */}
                            <TabsContent value="questions" className="mt-0">
                                <div className="space-y-4">
                                    <div className="mb-4">
                                        <h3 className="text-sm font-medium text-gray-900">Booking Questions</h3>
                                        <p className="text-sm text-gray-500">
                                            Ask guests for additional information when they book.
                                        </p>
                                    </div>
                                    <QuestionEditor
                                        questions={watch('questions') || []}
                                        onChange={(qs) => setValue('questions', qs)}
                                    />
                                </div>
                            </TabsContent>
                        </div>

                        <DialogFooter className="pt-4 border-t mt-auto">
                            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                            <Button type="submit" disabled={loading}>{loading ? "Saving..." : "Save Event Type"}</Button>
                        </DialogFooter>
                    </Tabs>
                </form>
            </DialogContent>
        </Dialog>
    );
}
