"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useToast } from '@/components/ui/use-toast';
import api from '@/lib/api/client'; // Public client should handle auth gracefully or be separate
import { appointmentApi } from '@/lib/api/appointment';

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
}

export default function BookingPage() {
    const params = useParams();
    const username = params.username as string; // Ideally this matches userId or we resolve username -> userId
    // For MVP, assuming username IS userInput ID in the url.
    const userId = username;
    const router = useRouter();

    const { toast } = useToast();
    const [availability, setAvailability] = useState<Availability | null>(null);
    const [selectedDate, setSelectedDate] = useState<string>('');
    const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [loading, setLoading] = useState(true);
    const [bookedSlots, setBookedSlots] = useState<string[][]>([]);

    useEffect(() => {
        const fetchBookedSlots = async () => {
            if (!userId || !selectedDate) return;
            try {
                const slots = await appointmentApi.getBookedSlots(userId, selectedDate);
                setBookedSlots(slots);
            } catch (error) {
                console.error('Failed to fetch booked slots', error);
            }
        };
        fetchBookedSlots();
    }, [userId, selectedDate]);

    const isSlotAvailable = (timeStr: string) => {
        if (!selectedDate) return false;
        const slotStart = new Date(`${selectedDate}T${timeStr}:00`);
        const slotEnd = new Date(slotStart.getTime() + 60 * 60 * 1000); // Assuming 1 hour duration

        return !bookedSlots.some(([start, end]) => {
            const bookedStart = new Date(start);
            const bookedEnd = new Date(end);
            return (slotStart < bookedEnd && slotEnd > bookedStart);
        });
    };

    useEffect(() => {
        const fetchAvailability = async () => {
            try {
                const res = await api.get<Availability>(`/users/${userId}/availability`);
                setAvailability(res.data);
            } catch (error) {
                console.error('Failed to fetch availability', error);
            } finally {
                setLoading(false);
            }
        };
        if (userId) fetchAvailability();
    }, [userId]);

    const handleBook = async () => {
        if (!selectedSlot || !selectedDate) return;
        setLoading(true);
        try {
            const startDateTime = new Date(`${selectedDate}T${selectedSlot}:00`);

            await appointmentApi.createPublicBooking(userId, {
                guest_name: name,
                guest_email: email,
                start_time: startDateTime.toISOString(),
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            });

            toast({ title: "Booking Confirmed", description: "You will receive an email shortly." });

            // Clear form
            setName('');
            setEmail('');
            setSelectedSlot(null);
        } catch (error) {
            console.error(error);
            toast({ title: "Booking Failed", description: "Could not book this slot.", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    if (loading) return <div className="p-8 text-center">Loading booking page...</div>;
    if (!availability) return <div className="p-8 text-center">User not found or no availability set.</div>;

    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
            <Card className="w-full max-w-3xl grid md:grid-cols-2 overflow-hidden">
                <div className="bg-white p-6 border-r">
                    <div className="mb-6">
                        <h2 className="text-sm font-semibold text-gray-500 uppercase">Host</h2>
                        <h1 className="text-2xl font-bold">{userId}</h1>
                    </div>
                    <div className="space-y-4">
                        <h3 className="font-semibold">Select a Date & Time</h3>
                        <Input type="date" value={selectedDate} onChange={(e) => setSelectedDate(e.target.value)} />

                        {selectedDate && (
                            <div className="grid grid-cols-2 gap-2 mt-4">
                                {['09:00', '10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00'].map(time => {
                                    const available = isSlotAvailable(time);
                                    return (
                                        <Button
                                            key={time}
                                            variant={selectedSlot === time ? 'default' : 'outline'}
                                            onClick={() => available && setSelectedSlot(time)}
                                            disabled={!available}
                                            className={!available ? 'opacity-50 cursor-not-allowed bg-gray-100 text-gray-400' : ''}
                                        >
                                            {available ? time : 'Booked'}
                                        </Button>
                                    );
                                })}
                            </div>
                        )}
                    </div>
                </div>

                <div className="p-6">
                    <CardHeader className="px-0 pt-0">
                        <CardTitle>Enter Details</CardTitle>
                        <CardDescription>Confirm your appointment</CardDescription>
                    </CardHeader>
                    <CardContent className="px-0 space-y-4">
                        <div className="space-y-2">
                            <Label>Name</Label>
                            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="John Doe" />
                        </div>
                        <div className="space-y-2">
                            <Label>Email</Label>
                            <Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="john@example.com" />
                        </div>
                        <Button className="w-full mt-4" disabled={!selectedSlot || !name || !email} onClick={handleBook}>
                            Confirm Booking
                        </Button>
                    </CardContent>
                </div>
            </Card>
        </div>
    );
}
