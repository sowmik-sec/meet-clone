"use client";

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '../../../components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select';
import { appointmentApi } from '@/lib/api/appointment';
// import { useAuth } from '@/hooks/useAuth';

import { Navigation } from '@/components/ui/Navigation';

export default function NewAppointmentPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(false);

    const [formData, setFormData] = useState({
        title: '',
        description: '',
        date: new Date().toISOString().split('T')[0],
        startTime: '09:00',
        endTime: '10:00',
        meetingType: 'meeting' as 'meeting' | 'webinar',
        guestId: '',
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            const start = new Date(`${formData.date}T${formData.startTime}:00`);
            const end = new Date(`${formData.date}T${formData.endTime}:00`);

            await appointmentApi.createAppointment({
                title: formData.title,
                description: formData.description,
                start_time: start.toISOString(),
                end_time: end.toISOString(),
                meeting_type: formData.meetingType,
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                guest_id: formData.guestId,
            });

            router.push('/schedule');
        } catch (error) {
            console.error('Failed to create appointment:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <Navigation />
            <div className="max-w-2xl mx-auto px-4 py-8">
                <Card>
                    <CardHeader>
                        <CardTitle>Schedule New Appointment</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-6">
                            <div className="space-y-2">
                                <Label htmlFor="title">Title</Label>
                                <Input
                                    id="title"
                                    required
                                    value={formData.title}
                                    onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                                    placeholder="Weekly Team Sync"
                                />
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="type">Meeting Type</Label>
                                    <Select
                                        value={formData.meetingType}
                                        onValueChange={(val: 'meeting' | 'webinar') => setFormData({ ...formData, meetingType: val })}
                                    >
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="meeting">Meeting</SelectItem>
                                            <SelectItem value="webinar">Webinar</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="guest">Guest User ID (Optional)</Label>
                                    <Input
                                        id="guest"
                                        value={formData.guestId}
                                        onChange={(e) => setFormData({ ...formData, guestId: e.target.value })}
                                        placeholder="User ID"
                                    />
                                </div>
                            </div>

                            <div className="grid grid-cols-3 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="date">Date</Label>
                                    <Input
                                        id="date"
                                        type="date"
                                        required
                                        value={formData.date}
                                        onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="start">Start Time</Label>
                                    <Input
                                        id="start"
                                        type="time"
                                        required
                                        value={formData.startTime}
                                        onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="end">End Time</Label>
                                    <Input
                                        id="end"
                                        type="time"
                                        required
                                        value={formData.endTime}
                                        onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="description">Description (Optional)</Label>
                                <Textarea
                                    id="description"
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    placeholder="Agenda details..."
                                />
                            </div>

                            <div className="flex gap-4 pt-4">
                                <Button type="button" variant="outline" className="flex-1" onClick={() => router.back()}>
                                    Cancel
                                </Button>
                                <Button type="submit" className="flex-1" disabled={loading}>
                                    {loading ? 'Scheduling...' : 'Schedule Appointment'}
                                </Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
