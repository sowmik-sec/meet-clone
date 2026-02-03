"use client";

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useToast } from '@/components/ui/use-toast';
import api from '@/lib/api/client'; // Public client should handle auth gracefully or be separate

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

    const { toast } = useToast();
    const [availability, setAvailability] = useState<Availability | null>(null);
    const [selectedDate, setSelectedDate] = useState<string>('');
    const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [loading, setLoading] = useState(true);

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
        // In Phase 3, this will create an appointment.
        // For now, just a toast.
        toast({ title: "Booking Request Sent", description: "This feature will be connected in Phase 3." });
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
                                {/* Mock slots for now based on schedule logic would go here */}
                                <Button variant={selectedSlot === '09:00' ? 'default' : 'outline'} onClick={() => setSelectedSlot('09:00')}>09:00 AM</Button>
                                <Button variant={selectedSlot === '10:00' ? 'default' : 'outline'} onClick={() => setSelectedSlot('10:00')}>10:00 AM</Button>
                                <Button variant={selectedSlot === '11:00' ? 'default' : 'outline'} onClick={() => setSelectedSlot('11:00')}>11:00 AM</Button>
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
