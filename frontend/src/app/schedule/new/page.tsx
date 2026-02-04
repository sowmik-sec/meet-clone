"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { format } from "date-fns";
import { ChevronsUpDown, Check, ChevronDown, ChevronUp, User, AlignLeft, CalendarIcon, Clock } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Navigation } from "@/components/ui/Navigation";
import { appointmentApi } from "@/lib/api/appointment";
import { toast } from "sonner";

const formSchema = z.object({
    title: z.string().min(1, "Title is required"),
    description: z.string().optional(),
    date: z.date(),
    startTime: z.string().min(1, "Start time is required"),
    endTime: z.string().min(1, "End time is required"),
    meetingType: z.enum(["meeting", "webinar"]),
    guestId: z.string().optional(),
});

function generateTimeOptions() {
    const options = [];
    for (let i = 0; i < 24; i++) {
        for (let j = 0; j < 60; j += 15) {
            const hour = i.toString().padStart(2, "0");
            const minute = j.toString().padStart(2, "0");
            options.push(`${hour}:${minute}`);
        }
    }
    return options;
}

const timeOptions = generateTimeOptions();

export default function NewAppointmentPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(false);

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            title: "",
            description: "",
            date: new Date(),
            startTime: "09:00",
            endTime: "09:30",
            meetingType: "meeting",
            guestId: "",
        },
    });

    async function onSubmit(values: z.infer<typeof formSchema>) {
        setLoading(true);
        try {
            // Combine date and time
            const dateStr = format(values.date, "yyyy-MM-dd");
            const startDateTime = new Date(`${dateStr}T${values.startTime}:00`);
            const endDateTime = new Date(`${dateStr}T${values.endTime}:00`);

            // Validate end time is after start time
            if (endDateTime <= startDateTime) {
                form.setError("endTime", {
                    type: "manual",
                    message: "End time must be after start time",
                });
                setLoading(false);
                return;
            }

            await appointmentApi.createAppointment({
                title: values.title,
                description: values.description,
                start_time: startDateTime.toISOString(),
                end_time: endDateTime.toISOString(),
                meeting_type: values.meetingType,
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                guest_id: values.guestId,
            });

            toast.success("Appointment scheduled successfully");
            router.push("/schedule");
        } catch (error) {
            console.error("Failed to create appointment:", error);
            toast.error("Failed to schedule appointment. Please try again.");
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="min-h-screen bg-gray-50/50">
            <Navigation />
            <div className="container mx-auto max-w-5xl px-4 py-8">
                <div className="mb-8">
                    <h1 className="text-3xl font-bold tracking-tight text-gray-900">Schedule Appointment</h1>
                    <p className="text-gray-500 mt-2">Create a new meeting or webinar manually.</p>
                </div>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)}>
                        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
                            {/* Left Column: Details */}
                            <div className="lg:col-span-7 space-y-6">
                                <Card className="border-none shadow-md">
                                    <CardHeader>
                                        <CardTitle>Details</CardTitle>
                                        <CardDescription>Enter the basic info for your meeting.</CardDescription>
                                    </CardHeader>
                                    <CardContent className="space-y-4">
                                        <FormField
                                            control={form.control}
                                            name="title"
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel>Title</FormLabel>
                                                    <FormControl>
                                                        <Input placeholder="e.g. Project Kickoff" {...field} className="text-lg font-medium" />
                                                    </FormControl>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />

                                        <FormField
                                            control={form.control}
                                            name="meetingType"
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel>Meeting Type</FormLabel>
                                                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                        <FormControl>
                                                            <SelectTrigger>
                                                                <SelectValue placeholder="Select type" />
                                                            </SelectTrigger>
                                                        </FormControl>
                                                        <SelectContent>
                                                            <SelectItem value="meeting">
                                                                <div className="flex items-center">
                                                                    <User className="mr-2 h-4 w-4" />
                                                                    <span>Standard Meeting</span>
                                                                </div>
                                                            </SelectItem>
                                                            <SelectItem value="webinar">
                                                                <div className="flex items-center">
                                                                    <AlignLeft className="mr-2 h-4 w-4" />
                                                                    <span>Webinar</span>
                                                                </div>
                                                            </SelectItem>
                                                        </SelectContent>
                                                    </Select>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />

                                        <FormField
                                            control={form.control}
                                            name="guestId"
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel>Guest User ID (Optional)</FormLabel>
                                                    <FormControl>
                                                        <Input placeholder="Directly invite a registered user by ID" {...field} />
                                                    </FormControl>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />

                                        <FormField
                                            control={form.control}
                                            name="description"
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel>Description</FormLabel>
                                                    <FormControl>
                                                        <Textarea
                                                            placeholder="Add agenda, links, or notes..."
                                                            className="min-h-[120px]"
                                                            {...field}
                                                        />
                                                    </FormControl>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />
                                    </CardContent>
                                </Card>
                            </div>

                            {/* Right Column: Date & Time */}
                            <div className="lg:col-span-5 space-y-6">
                                <Card className="border-none shadow-md overflow-hidden">
                                    <CardHeader className="bg-blue-50/50 border-b border-blue-100/50 pb-4">
                                        <CardTitle className="flex items-center text-blue-900">
                                            <CalendarIcon className="mr-2 h-5 w-5" />
                                            Date & Time
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent className="p-0">
                                        <div className="p-4 flex justify-center bg-white">
                                            <FormField
                                                control={form.control}
                                                name="date"
                                                render={({ field }) => (
                                                    <FormItem className="flex flex-col">
                                                        <FormControl>
                                                            <Calendar
                                                                mode="single"
                                                                selected={field.value}
                                                                onSelect={field.onChange}
                                                                disabled={(date) =>
                                                                    date < new Date(new Date().setHours(0, 0, 0, 0))
                                                                }
                                                                initialFocus
                                                                className="rounded-md border shadow-sm"
                                                            />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </FormItem>
                                                )}
                                            />
                                        </div>

                                        <div className="p-6 border-t bg-gray-50/50 space-y-4">
                                            <div className="flex gap-4">
                                                <FormField
                                                    control={form.control}
                                                    name="startTime"
                                                    render={({ field }) => (
                                                        <FormItem className="flex-1">
                                                            <FormLabel>Start Time</FormLabel>
                                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                                <FormControl>
                                                                    <SelectTrigger>
                                                                        <SelectValue placeholder="Start" />
                                                                    </SelectTrigger>
                                                                </FormControl>
                                                                <SelectContent className="max-h-[200px]">
                                                                    {timeOptions.map((time) => (
                                                                        <SelectItem key={`start-${time}`} value={time}>
                                                                            {time}
                                                                        </SelectItem>
                                                                    ))}
                                                                </SelectContent>
                                                            </Select>
                                                            <FormMessage />
                                                        </FormItem>
                                                    )}
                                                />

                                                <FormField
                                                    control={form.control}
                                                    name="endTime"
                                                    render={({ field }) => (
                                                        <FormItem className="flex-1">
                                                            <FormLabel>End Time</FormLabel>
                                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                                <FormControl>
                                                                    <SelectTrigger>
                                                                        <SelectValue placeholder="End" />
                                                                    </SelectTrigger>
                                                                </FormControl>
                                                                <SelectContent className="max-h-[200px]">
                                                                    {timeOptions.map((time) => (
                                                                        <SelectItem key={`end-${time}`} value={time}>
                                                                            {time}
                                                                        </SelectItem>
                                                                    ))}
                                                                </SelectContent>
                                                            </Select>
                                                            <FormMessage />
                                                        </FormItem>
                                                    )}
                                                />
                                            </div>

                                            <div className="flex items-center text-sm text-gray-500 pt-2">
                                                <Clock className="mr-2 h-4 w-4" />
                                                <span>Timezone: {Intl.DateTimeFormat().resolvedOptions().timeZone}</span>
                                            </div>
                                        </div>
                                    </CardContent>
                                    <div className="p-6 border-t bg-gray-50/50 grid grid-cols-2 gap-4">
                                        <Button
                                            type="button"
                                            variant="outline"
                                            className="w-full"
                                            onClick={() => router.back()}
                                        >
                                            Cancel
                                        </Button>
                                        <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-700" disabled={loading}>
                                            {loading ? "Scheduling..." : "Create Appointment"}
                                        </Button>
                                    </div>
                                </Card>
                            </div>
                        </div>
                    </form>
                </Form>
            </div>
        </div>
    );
}
