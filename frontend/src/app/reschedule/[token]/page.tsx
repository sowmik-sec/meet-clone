"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { useToast } from '@/components/ui/use-toast';
import api from '@/lib/api/client';
import { eventTypesApi } from '@/lib/api/event-types';
import { appointmentApi } from '@/lib/api/appointment';
import { HostInfoCard } from '@/components/booking/host-info-card';
import { BookingCalendar } from '@/components/booking/booking-calendar';
import { TimeSlotGrid } from '@/components/booking/time-slot-grid';
import { ArrowLeft, Calendar as CalendarIcon, Clock, AlertTriangle } from 'lucide-react';
import { format } from 'date-fns';
import { EventType } from '@/types/event-type';
import { Appointment } from '@/types/appointment';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

interface TimeSlot {
    start: string;
    end: string;
}

interface DayAvailability {
    day: number;
    is_enabled: boolean;
    slots: TimeSlot[];
}

interface Availability {
    user_id: string;
    schedule: DayAvailability[];
    timezone: string;
    is_accepting_bookings: boolean;
}

export default function ReschedulePage() {
    const params = useParams();
    const token = params.token as string;
    const router = useRouter();

    const { toast } = useToast();
    const [appointment, setAppointment] = useState<Appointment | null>(null);
    const [availability, setAvailability] = useState<Availability | null>(null);
    const [eventType, setEventType] = useState<EventType | null>(null);
    const [selectedDate, setSelectedDate] = useState<Date | undefined>(undefined);
    const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const [bookedSlots, setBookedSlots] = useState<string[][]>([]);
    const [error, setError] = useState<string | null>(null);
    const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
    const [isCancelling, setIsCancelling] = useState(false);

    // View state: 'calendar' or 'confirm'
    const [view, setView] = useState<'calendar' | 'confirm'>('calendar');

    useEffect(() => {
        const loadData = async () => {
            setLoading(true);
            try {
                // 1. Fetch appointment by token
                const appt = await appointmentApi.getAppointmentByRescheduleToken(token);
                setAppointment(appt);

                if (appt && appt.event_type_id) {
                    // 2. Fetch Event Type
                    // We need user ID to fetch public event types? 
                    // Or we need a way to get event type by ID PUBLICLY?
                    // Currently `eventTypesApi.listPublic(userId)` lists all.
                    // We can use appt.host_id.
                    const types = await eventTypesApi.listPublic(appt.host_id);
                    const et = types.find(t => t.id === appt.event_type_id);
                    if (et) {
                        setEventType(et);
                    }

                    // 3. Fetch Availability
                    const res = await api.get<Availability>(`/users/${appt.host_id}/availability`);
                    setAvailability(res.data);
                }
            } catch (error: any) {
                console.error("Failed to load reschedule data", error);
                setError(error.response?.data?.error || "Invalid or expired reschedule link.");
            } finally {
                setLoading(false);
            }
        };
        if (token) loadData();
    }, [token]);

    useEffect(() => {
        const fetchBookedSlots = async () => {
            if (!appointment?.host_id || !selectedDate) return;
            try {
                const dateStr = format(selectedDate, 'yyyy-MM-dd');
                const response = await appointmentApi.getBookedSlots(appointment.host_id, dateStr);
                setBookedSlots(response.busy_slots);
            } catch (error) {
                console.error('Failed to fetch booked slots', error);
            }
        };
        fetchBookedSlots();
    }, [appointment?.host_id, selectedDate]);

    const getAvailableSlots = () => {
        if (!availability || !selectedDate || !eventType) return [];

        const date = selectedDate;
        const dayIndex = date.getDay(); // 0 = Sunday
        const dayConfig = availability.schedule.find(d => d.day === dayIndex);

        if (!dayConfig || !dayConfig.is_enabled) return [];

        const slots: string[] = [];
        const dateStr = format(date, 'yyyy-MM-dd');
        const durationMin = eventType.duration;

        dayConfig.slots.forEach(slotRange => {
            const [startHour, startMin] = slotRange.start.split(':').map(Number);
            const [endHour, endMin] = slotRange.end.split(':').map(Number);

            let current = new Date(dateStr + 'T00:00:00');
            current.setHours(startHour, startMin, 0, 0);

            let end = new Date(dateStr + 'T00:00:00');
            end.setHours(endHour, endMin, 0, 0);

            const now = new Date();
            // Consider MinRescheduleNotice? 
            // The backend enforces it, frontend should ideally reflect it too.
            // But for simplicity, let's just use "now". 
            // If MinRescheduleNotice > 0, we should maybe add that to "now".
            // But that applies to the OLD appointment time vs now.
            // For the NEW time, it just needs to be in future (and available).

            const isToday = date.toDateString() === now.toDateString();

            while (current < end) {
                const nextSlot = new Date(current.getTime() + durationMin * 60 * 1000);

                if (isToday && current <= now) {
                    const step = durationMin;
                    current = new Date(current.getTime() + step * 60 * 1000);
                    continue;
                }

                if (nextSlot <= end) {
                    const timeString = current.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
                    if (!slots.includes(timeString)) {
                        slots.push(timeString);
                    }
                }
                // Use duration + buffers as step
                const totalBuffer = (eventType.buffer_before || 0) + (eventType.buffer_after || 0);
                const step = durationMin + totalBuffer;
                current = new Date(current.getTime() + step * 60 * 1000);
            }
        });

        return slots;
    };

    const isSlotAvailable = (timeStr: string, date: Date) => {
        if (!eventType) return false;
        const dateStr = format(date, 'yyyy-MM-dd');
        const slotStart = new Date(`${dateStr}T${timeStr}:00`);
        const slotEnd = new Date(slotStart.getTime() + eventType.duration * 60 * 1000);

        // Include buffer before and after for the new slot
        const bufferBefore = (eventType.buffer_before || 0) * 60 * 1000;
        const bufferAfter = (eventType.buffer_after || 0) * 60 * 1000;
        const bufferedStart = new Date(slotStart.getTime() - bufferBefore);
        const bufferedEnd = new Date(slotEnd.getTime() + bufferAfter);

        return !bookedSlots?.some(([start, end]) => {
            const bookedStart = new Date(start);
            const bookedEnd = new Date(end);
            // Check if the buffered slot overlaps with any booked slot
            return (bufferedStart < bookedEnd && bufferedEnd > bookedStart);
        });
    };

    const handleSlotSelect = (slot: string) => {
        setSelectedSlot(slot);
        setView('confirm');
    };

    const handleBackToCalendar = () => {
        setView('calendar');
        setSelectedSlot(null);
    };

    const handleCancelAppointment = async () => {
        setIsCancelling(true);
        try {
            await appointmentApi.cancelAppointmentByToken(token);
            toast({
                title: "Appointment Cancelled",
                description: "This appointment has been successfully cancelled.",
            });
            setError("cancelled"); // Show cancelled state
            setCancelDialogOpen(false);
        } catch (error: any) {
            console.error(error);
            toast({
                title: "Cancellation Failed",
                description: error.response?.data?.error || "Could not cancel appointment.",
                variant: "destructive"
            });
            setIsCancelling(false);
        }
    };

    const handleReschedule = async () => {
        if (!selectedSlot || !selectedDate) return;
        setLoading(true);
        try {
            const dateStr = format(selectedDate, 'yyyy-MM-dd');
            // Construct ISO string with timezone??
            // The API expects time.Time which handles ISO.
            // We should use the local time selected but send it as ISO.
            const startDateTime = new Date(`${dateStr}T${selectedSlot}:00`);

            await appointmentApi.rescheduleAppointment(token, startDateTime.toISOString());

            toast({ title: "Rescheduled!", description: "Your appointment has been updated." });

            // Redirect to a success page or show success state?
            // Maybe redirect to the appointment details / dashboard / or just show success here.
            // Let's show a success state here.

            // For now, reload to show updated time? Or redirect to generic success?
            // Redirecting to home or dashboard might be confusing for guest.
            // Let's just update local state to show "Success".
            router.push('/'); // Or maybe stay?
            // Actually, let's just force a reload or fetch new appointment data
            //  window.location.reload();
            // Best UX: Show success message/card replacing the form.
            setError("success"); // Hack to show success view

        } catch (error: any) {
            console.error(error);
            toast({ title: "Reschedule Failed", description: error.response?.data?.error || "Could not reschedule.", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    if (loading) return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="animate-pulse flex flex-col items-center">
                <div className="h-12 w-12 bg-gray-200 rounded-full mb-4"></div>
                <div className="h-4 w-32 bg-gray-200 rounded"></div>
            </div>
        </div>
    );

    if (error === "success") {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
                <Card className="max-w-md w-full text-center p-8 bg-green-50 border-green-100">
                    <div className="flex justify-center mb-4">
                        <div className="h-12 w-12 bg-green-100 rounded-full flex items-center justify-center">
                            <Clock className="w-6 h-6 text-green-600" />
                        </div>
                    </div>
                    <h2 className="text-xl font-bold text-green-900 mb-2">Appointment Rescheduled!</h2>
                    <p className="text-green-700 mb-6">
                        Your appointment has been successfully updated to {selectedDate && format(selectedDate, 'MMMM d, yyyy')} at {selectedSlot}.
                        Check your email for confirmation.
                    </p>
                    <Button onClick={() => router.push('/')} variant="outline" className="border-green-200 text-green-800 hover:bg-green-100">
                        Go Home
                    </Button>
                </Card>
            </div>
        )
    }

    if (error === "cancelled") {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
                <Card className="max-w-md w-full text-center p-8 bg-red-50 border-red-100">
                    <div className="flex justify-center mb-4">
                        <div className="h-12 w-12 bg-red-100 rounded-full flex items-center justify-center">
                            <AlertTriangle className="w-6 h-6 text-red-600" />
                        </div>
                    </div>
                    <h2 className="text-xl font-bold text-red-900 mb-2">Appointment Cancelled</h2>
                    <p className="text-red-700 mb-6">
                        This appointment has been cancelled as requested.
                    </p>
                    <Button onClick={() => router.push('/')} variant="outline" className="border-red-200 text-red-800 hover:bg-red-100">
                        Go Home
                    </Button>
                </Card>
            </div>
        )
    }

    if (error || !appointment) return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
            <Card className="max-w-md w-full text-center p-8">
                <AlertTriangle className="w-12 h-12 text-yellow-500 mx-auto mb-4" />
                <h2 className="text-xl font-semibold text-gray-900 mb-2">Unavailable</h2>
                <p className="text-gray-500">{error || "Appointment not found or expired."}</p>
            </Card>
        </div>
    );

    // If bookings paused?
    if (availability && !availability.is_accepting_bookings) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="text-center max-w-md p-6 bg-white rounded-lg shadow-sm border border-gray-100">
                    <h2 className="text-xl font-semibold text-gray-900 mb-2">Bookings Paused</h2>
                    <p className="text-gray-500">The host is not accepting reschedule requests at this time.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
            <Card className="w-full max-w-5xl shadow-xl overflow-hidden min-h-[600px] flex md:flex-row flex-col">
                {/* Info Sidebar */}
                <div className="w-full md:w-1/3 bg-white border-b md:border-b-0 md:border-r border-gray-100 p-6">
                    <div className="mb-6">
                        <span className="bg-blue-100 text-blue-700 text-xs font-semibold px-2 py-1 rounded uppercase tracking-wide">
                            Rescheduling
                        </span>
                    </div>

                    <h2 className="text-2xl font-bold text-gray-900 mb-2">{appointment.title}</h2>
                    <p className="text-gray-500 mb-6">{eventType?.description}</p>

                    <div className="space-y-4">
                        <div className="flex items-start gap-3 text-gray-600">
                            <Clock className="w-5 h-5 mt-0.5 text-gray-400" />
                            <div>
                                <p className="font-medium text-gray-900">Current Time</p>
                                <p className="text-sm">
                                    {format(new Date(appointment.start_time), 'EEEE, MMMM d, yyyy')}
                                    <br />
                                    {format(new Date(appointment.start_time), 'st')} - {format(new Date(appointment.end_time), 'p')}
                                </p>
                            </div>
                        </div>

                        <div className="flex items-center gap-3 text-gray-600">
                            <Clock className="w-5 h-5 text-gray-400" />
                            <span>{eventType?.duration || 30} mins</span>
                        </div>

                        <div className="pt-6 mt-6 border-t border-gray-100">
                            <Button
                                variant="outline"
                                className="w-full text-red-600 border-red-200 hover:bg-red-50 hover:text-red-700"
                                onClick={() => setCancelDialogOpen(true)}
                            >
                                Cancel Appointment
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Main Content */}
                <div className="flex-1 bg-white p-6 md:p-8">
                    {view === 'calendar' ? (
                        <div className="h-full flex flex-col md:flex-row gap-8 animate-in fade-in duration-500">
                            <div className="flex-1 flex flex-col items-center">
                                <h3 className="text-lg font-semibold mb-4 self-start">Select New Date</h3>
                                <BookingCalendar
                                    selectedDate={selectedDate}
                                    onSelectDate={setSelectedDate}
                                />
                            </div>

                            <div className={`w-full md:w-64 border-l pl-0 md:pl-8 transition-all duration-300 ${!selectedDate ? 'opacity-50 pointer-events-none' : 'opacity-100'}`}>
                                <h3 className="text-lg font-semibold mb-4">Available Time</h3>
                                <TimeSlotGrid
                                    date={selectedDate}
                                    slots={getAvailableSlots()}
                                    selectedSlot={selectedSlot}
                                    onSelectSlot={handleSlotSelect}
                                    isSlotAvailable={isSlotAvailable}
                                />
                            </div>
                        </div>
                    ) : (
                        <div className="h-full max-w-md mx-auto animate-in slide-in-from-right-8 duration-300">
                            <Button
                                variant="ghost"
                                onClick={handleBackToCalendar}
                                className="mb-6 -ml-4 text-gray-500 hover:text-gray-900"
                            >
                                <ArrowLeft className="w-4 h-4 mr-2" />
                                Back to Calendar
                            </Button>

                            <div className="mb-8 p-4 bg-blue-50 rounded-lg border border-blue-100">
                                <h3 className="font-semibold text-blue-900 mb-2">New Time Selection</h3>
                                <div className="space-y-2 text-sm text-blue-800">
                                    <div className="flex items-center gap-2">
                                        <CalendarIcon className="w-4 h-4" />
                                        <span>{selectedDate && format(selectedDate, 'EEEE, MMMM d, yyyy')}</span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Clock className="w-4 h-4" />
                                        <span>{selectedSlot} - {selectedSlot && eventType && format(new Date(`2000-01-01T${selectedSlot}:00`).getTime() + eventType.duration * 60 * 1000, 'HH:mm')}</span>
                                    </div>
                                </div>
                            </div>

                            <Button
                                className="w-full h-11 text-base bg-blue-600 hover:bg-blue-700"
                                onClick={handleReschedule}
                            >
                                Confirm Reschedule
                            </Button>
                        </div>
                    )}
                </div>
            </Card>

            <Dialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Cancel Appointment</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to cancel this appointment? This action cannot be undone.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCancelDialogOpen(false)}>Keep Appointment</Button>
                        <Button variant="destructive" onClick={handleCancelAppointment} disabled={isCancelling}>
                            {isCancelling ? "Cancelling..." : "Yes, Cancel Appointment"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
