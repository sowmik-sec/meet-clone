"use client";

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { useAuth } from '@/hooks/useAuth';
import api from '@/lib/api/client';
import { useToast } from '@/components/ui/use-toast';
import { Navigation } from '@/components/ui/Navigation';
import { Copy, Check } from 'lucide-react';

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

const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export default function AvailabilityPage() {
    const { isAuthenticated, user } = useAuth();
    const { toast } = useToast();
    const [availability, setAvailability] = useState<Availability | null>(null);
    const [loading, setLoading] = useState(true);
    const [copied, setCopied] = useState(false);

    useEffect(() => {
        const fetchAvailability = async () => {
            try {
                const res = await api.get<Availability>('/availability');
                setAvailability(res.data);
            } catch (error) {
                console.error('Failed to fetch availability', error);
            } finally {
                setLoading(false);
            }
        };
        if (isAuthenticated) fetchAvailability();
    }, [isAuthenticated]);

    const handleSave = async () => {
        if (!availability) return;
        try {
            await api.post('/availability', availability);
            toast({ title: 'Success', description: 'Availability saved successfully.' });
        } catch (error) {
            console.error('Failed to save', error);
            toast({ title: 'Error', description: 'Failed to save availability.', variant: 'destructive' });
        }
    };

    const toggleDay = (dayIndex: number) => {
        if (!availability) return;
        const newSchedule = [...availability.schedule];
        newSchedule[dayIndex].is_enabled = !newSchedule[dayIndex].is_enabled;
        setAvailability({ ...availability, schedule: newSchedule });
    };

    const updateSlot = (dayIndex: number, slotIndex: number, field: 'start' | 'end', value: string) => {
        if (!availability) return;
        const newSchedule = [...availability.schedule];
        newSchedule[dayIndex].slots[slotIndex][field] = value;
        setAvailability({ ...availability, schedule: newSchedule });
    };

    const copyBookingLink = () => {
        if (!user?.id) return;
        const link = `${window.location.origin}/b/${user.id}`;
        navigator.clipboard.writeText(link);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
        toast({ title: 'Copied', description: 'Booking link copied to clipboard.' });
    };

    if (!availability && !loading) return <div>Failed to load</div>;
    if (loading) return <div>Loading...</div>;

    return (
        <div className="min-h-screen bg-gray-50">
            <Navigation />
            <div className="max-w-4xl mx-auto px-4 py-8">
                <div className="flex justify-between items-center mb-6">
                    <h1 className="text-2xl font-bold">Set Your Availability</h1>
                    <Button onClick={handleSave}>Save Changes</Button>
                </div>

                <div className="flex justify-between items-center mb-6">
                    <div>
                        <CardTitle>Availability</CardTitle>
                        <CardDescription>Set your weekly schedule</CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="outline" onClick={copyBookingLink}>
                            {copied ? <Check className="w-4 h-4 mr-2" /> : <Copy className="w-4 h-4 mr-2" />}
                            Copy Booking Link
                        </Button>
                        <Button variant="outline" onClick={() => window.location.href = `${process.env.NEXT_PUBLIC_API_URL}/auth/google`}>
                            Sync Google Calendar
                        </Button>
                    </div>
                </div>

                <div className="space-y-4">
                    {availability?.schedule.map((day, index) => (
                        <Card key={index}>
                            <CardContent className="flex items-center gap-4 py-4">
                                <div className="w-32 flex items-center gap-2">
                                    <input
                                        type="checkbox"
                                        checked={day.is_enabled}
                                        onChange={() => toggleDay(index)}
                                        className="h-4 w-4"
                                    />
                                    <span className="font-medium">{DAYS[day.day]}</span>
                                </div>

                                {day.is_enabled ? (
                                    <div className="flex-1 space-y-2">
                                        {day.slots.map((slot, slotIndex) => (
                                            <div key={slotIndex} className="flex gap-2 items-center">
                                                <Input
                                                    type="time"
                                                    value={slot.start}
                                                    onChange={(e) => updateSlot(index, slotIndex, 'start', e.target.value)}
                                                    className="w-32"
                                                />
                                                <span>-</span>
                                                <Input
                                                    type="time"
                                                    value={slot.end}
                                                    onChange={(e) => updateSlot(index, slotIndex, 'end', e.target.value)}
                                                    className="w-32"
                                                />
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-gray-400">Unavailable</div>
                                )}
                            </CardContent>
                        </Card>
                    ))}
                </div>
            </div>
        </div>
    );
}
