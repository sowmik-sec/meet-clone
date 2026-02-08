"use client";

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useToast } from '@/components/ui/use-toast';
import api from '@/lib/api/client';
import { eventTypesApi } from '@/lib/api/event-types';
import { appointmentApi } from '@/lib/api/appointment';
import { HostInfoCard } from '@/components/booking/host-info-card';
import { BookingCalendar } from '@/components/booking/booking-calendar';
import { TimeSlotGrid } from '@/components/booking/time-slot-grid';
import { ArrowLeft, Calendar as CalendarIcon, Clock } from 'lucide-react';
import { format } from 'date-fns';
import { EventType } from '@/types/event-type';

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

export default function BookingPage() {
    const params = useParams();
    const username = params.username as string;
    const eventSlug = params.eventSlug as string;
    const userId = username;

    const { toast } = useToast();
    const [availability, setAvailability] = useState<Availability | null>(null);
    const [eventType, setEventType] = useState<EventType | null>(null);
    const [selectedDate, setSelectedDate] = useState<Date | undefined>(undefined);
    const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [loading, setLoading] = useState(true);
    const [bookedSlots, setBookedSlots] = useState<string[][]>([]);
    const [partialSlots, setPartialSlots] = useState<Record<string, number>>({});

    // View state: 'calendar' or 'form'
    const [view, setView] = useState<'calendar' | 'form'>('calendar');

    useEffect(() => {
        const loadData = async () => {
            setLoading(true);
            try {
                // Fetch public event types to find the current one
                const types = await eventTypesApi.listPublic(userId);
                const currentType = types.find(t => t.slug === eventSlug);

                if (currentType) {
                    setEventType(currentType);
                    // Only fetch availability if we found the event type
                    const res = await api.get<Availability>(`/users/${userId}/availability`);
                    setAvailability(res.data);
                }
            } catch (error) {
                console.error("Failed to load booking data", error);
            } finally {
                setLoading(false);
            }
        };
        if (userId && eventSlug) loadData();
    }, [userId, eventSlug]);

    useEffect(() => {
        const fetchBookedSlots = async () => {
            if (!userId || !selectedDate) return;
            try {
                const dateStr = format(selectedDate, 'yyyy-MM-dd');
                const response = await appointmentApi.getBookedSlots(userId, dateStr, eventType?.id);
                setBookedSlots(response.busy_slots || []);
                setPartialSlots(response.partial_slots || {});
            } catch (error) {
                console.error('Failed to fetch booked slots', error);
            }
        };
        fetchBookedSlots();
    }, [userId, selectedDate, eventType]);

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
            const isToday = date.toDateString() === now.toDateString();

            while (current < end) {
                const nextSlot = new Date(current.getTime() + durationMin * 60 * 1000);

                // Prevent past time selection if date is today
                if (isToday && current <= now) {
                    // Try to move to next interval (e.g. 15 or 30 mins step?)
                    // For now, assume step equals duration or 30 mins default step?
                    // Cal.com uses step (e.g. 15min slots for 30min meeting).
                    // Let's stick to Duration as Step for MVP simplicity
                    current = new Date(current.getTime() + (durationMin < 30 ? 15 : 30) * 60 * 1000);
                    continue;
                }

                if (nextSlot <= end) {
                    const timeString = current.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
                    if (!slots.includes(timeString)) {
                        slots.push(timeString);
                    }
                }
                // Increment by 30 mins or duration? 
                // Typically you want start times every 15/30/60 mins irrelevant of duration.
                // Let's use 30 min step for now, or 15 if duration is small.
                // Use duration + buffers as step to ensure back-to-back booking availability
                // const step = durationMin;
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

    // Fetched in loadData
    // useEffect(() => {
    //     const fetchAvailability = async () => { ... }
    // }, [userId]);

    const handleSlotSelect = (slot: string) => {
        setSelectedSlot(slot);
        setView('form');
    };

    const handleBackToCalendar = () => {
        setView('calendar');
        setSelectedSlot(null);
    };

    const handleBook = async () => {
        if (!selectedSlot || !selectedDate) return;
        setLoading(true);
        try {
            const dateStr = format(selectedDate, 'yyyy-MM-dd');
            const startDateTime = new Date(`${dateStr}T${selectedSlot}:00`);

            const appt = await appointmentApi.createPublicBooking(userId, {
                guest_name: name,
                guest_email: email,
                start_time: startDateTime.toISOString(),
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                event_type_id: eventType?.id || '',
            });

            if (appt.status === 'pending') {
                toast({ title: "Booking Requested", description: "Your booking is pending approval. You will receive an email shortly." });
            } else {
                toast({ title: "Booking Confirmed", description: "You will receive an email shortly." });
            }

            // Reset
            setName('');
            setEmail('');
            setSelectedSlot(null);
            setView('calendar');
            setSelectedDate(undefined);
        } catch (error) {
            console.error(error);
            toast({ title: "Booking Failed", description: "Could not book this slot.", variant: "destructive" });
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

    if (!availability) return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="text-center">
                <h2 className="text-xl font-semibold text-gray-900">User not found</h2>
                <p className="text-gray-500">This user hasn't set up their availability yet.</p>
            </div>
        </div>
    );

    if (!availability.is_accepting_bookings) return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="text-center max-w-md p-6 bg-white rounded-lg shadow-sm border border-gray-100">
                <h2 className="text-xl font-semibold text-gray-900 mb-2">Bookings Paused</h2>
                <p className="text-gray-500">This user is not currently accepting new bookings. Please try again later.</p>
            </div>
        </div>
    );


    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
            <Card className="w-full max-w-5xl shadow-xl overflow-hidden min-h-[600px] flex md:flex-row flex-col">
                {/* Host Sidebar - Always visible on desktop */}
                <div className="w-full md:w-1/3 bg-white border-b md:border-b-0 md:border-r border-gray-100">
                    <HostInfoCard
                        hostId={userId}
                        hostName={userId} // TODO: Fetch real name
                        description={eventType ? eventType.description : "Welcome to my scheduling page."}
                        duration={eventType?.duration}
                        title={eventType?.title}
                    />
                </div>

                {/* Main Content Area */}
                <div className="flex-1 bg-white p-6 md:p-8">
                    {view === 'calendar' ? (
                        <div className="h-full flex flex-col md:flex-row gap-8 animate-in fade-in duration-500">
                            <div className="flex-1 flex flex-col items-center">
                                <h3 className="text-lg font-semibold mb-4 self-start">Select a Date</h3>
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
                                    partialSlots={partialSlots} // Pass partial slots map
                                    maxAttendees={eventType?.max_attendees || 1} // Pass max attendees
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

                            <div className="mb-8 p-4 bg-gray-50 rounded-lg border border-gray-100">
                                <h3 className="font-semibold text-gray-900 mb-2">Booking Summary</h3>
                                <div className="space-y-2 text-sm text-gray-600">
                                    <div className="flex items-center gap-2">
                                        <CalendarIcon className="w-4 h-4" />
                                        <span>{selectedDate && format(selectedDate, 'EEEE, MMMM d, yyyy')}</span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Clock className="w-4 h-4" />
                                        <span>{selectedSlot} - {selectedSlot && eventType && format(new Date(`2000-01-01T${selectedSlot}:00`).getTime() + eventType.duration * 60 * 1000, 'HH:mm')}</span>
                                    </div>
                                    <div className="text-xs text-gray-400 mt-2 pt-2 border-t">
                                        Timezone: {Intl.DateTimeFormat().resolvedOptions().timeZone}
                                    </div>
                                </div>
                            </div>

                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <Label htmlFor="name">Your Name</Label>
                                    <Input
                                        id="name"
                                        value={name}
                                        onChange={(e) => setName(e.target.value)}
                                        placeholder="John Doe"
                                        className="h-11"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="email">Email Address</Label>
                                    <Input
                                        id="email"
                                        type="email"
                                        value={email}
                                        onChange={(e) => setEmail(e.target.value)}
                                        placeholder="john@example.com"
                                        className="h-11"
                                    />
                                </div>
                                <Button
                                    className="w-full h-11 mt-4 text-base"
                                    disabled={!name || !email}
                                    onClick={handleBook}
                                >
                                    Confirm Booking
                                </Button>
                            </div>
                        </div>
                    )}
                </div>
            </Card>
        </div>
    );
}
