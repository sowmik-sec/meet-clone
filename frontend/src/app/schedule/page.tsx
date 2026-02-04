"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { appointmentApi } from '@/lib/api/appointment';
import { Appointment } from '@/types/appointment';
import { useAuth } from '@/hooks/useAuth';
import { Calendar, Clock, Video, Users, Plus } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

import { Navigation } from '@/components/ui/Navigation';

const doesMeetingAllowStart = (startTime: string, endTime: string) => {
    const now = new Date();
    const start = new Date(startTime);
    const end = new Date(endTime);
    // Allow start 15 mins before
    const windowStart = new Date(start.getTime() - 15 * 60 * 1000);

    return now >= windowStart && now <= end;
};

export default function SchedulePage() {
    const router = useRouter();
    const { isAuthenticated, hasHydrated } = useAuth();
    const { toast } = useToast();
    const [appointments, setAppointments] = useState<Appointment[]>([]);
    const [loading, setLoading] = useState(true);
    const [cancelId, setCancelId] = useState<string | null>(null);

    useEffect(() => {
        if (hasHydrated && !isAuthenticated) {
            router.push('/login');
        }
    }, [isAuthenticated, hasHydrated, router]);

    useEffect(() => {
        const fetchAppointments = async () => {
            try {
                const data = await appointmentApi.getAppointments({
                    start_time_after: new Date().toISOString(),
                });
                setAppointments(data);
            } catch (error) {
                console.error('Failed to fetch appointments:', error);
                toast({
                    title: "Error",
                    description: "Failed to load appointments.",
                    variant: "destructive",
                });
            } finally {
                setLoading(false);
            }
        };

        if (isAuthenticated) {
            fetchAppointments();
        }
    }, [isAuthenticated, toast]);

    const handleStartMeeting = async (id: string) => {
        try {
            const { room_id } = await appointmentApi.startAppointment(id);
            toast({
                title: "Starting Meeting",
                description: "Redirecting you to the room...",
            });
            router.push(`/room/${room_id}`);
        } catch (error) {
            console.error('Failed to start meeting:', error);
            toast({
                title: "Error",
                description: "Failed to start meeting. Please try again.",
                variant: "destructive",
            });
        }
    };

    const confirmCancel = async () => {
        if (!cancelId) return;
        try {
            await appointmentApi.cancelAppointment(cancelId);
            setAppointments(appointments.filter(a => a.id !== cancelId));
            toast({
                title: "Appointment Cancelled",
                description: "The appointment has been successfully cancelled.",
            });
        } catch (error) {
            console.error('Failed to cancel appointment:', error);
            toast({
                title: "Error",
                description: "Failed to cancel appointment. Please try again.",
                variant: "destructive",
            });
        } finally {
            setCancelId(null);
        }
    };

    if (!hasHydrated || loading) {
        return <div className="p-8">Loading schedule...</div>;
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <Navigation />
            <div className="max-w-4xl mx-auto px-4 py-8">
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold">Schedule</h1>
                        <p className="text-gray-500">Manage your upcoming meetings and webinars</p>
                    </div>
                    <Button onClick={() => router.push('/schedule/new')}>
                        <Plus className="w-4 h-4 mr-2" />
                        New Appointment
                    </Button>
                </div>

                <div className="space-y-4">
                    {appointments.length === 0 ? (
                        <Card>
                            <CardContent className="p-8 text-center text-gray-500">
                                No upcoming appointments found.
                            </CardContent>
                        </Card>
                    ) : (
                        appointments.map((appt) => (
                            <Card key={appt.id}>
                                <CardContent className="p-6 flex justify-between items-center">
                                    <div>
                                        <h3 className="text-xl font-semibold mb-2">{appt.title}</h3>
                                        <div className="flex gap-4 text-sm text-gray-500">
                                            <div className="flex items-center">
                                                <Calendar className="w-4 h-4 mr-1" />
                                                {new Date(appt.start_time).toLocaleDateString()}
                                            </div>
                                            <div className="flex items-center">
                                                <Clock className="w-4 h-4 mr-1" />
                                                {new Date(appt.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} -
                                                {new Date(appt.end_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                            </div>
                                            <div className="flex items-center capitalize">
                                                {appt.meeting_type === 'webinar' ? <Users className="w-4 h-4 mr-1" /> : <Video className="w-4 h-4 mr-1" />}
                                                {appt.meeting_type}
                                            </div>
                                            <span className={`px-2 py-0.5 rounded text-xs capitalize ${appt.status === 'confirmed' ? 'bg-green-100 text-green-800' :
                                                appt.status === 'cancelled' ? 'bg-red-100 text-red-800' :
                                                    'bg-yellow-100 text-yellow-800'
                                                }`}>
                                                {appt.status}
                                            </span>
                                        </div>
                                        {appt.description && <p className="mt-2 text-gray-600">{appt.description}</p>}
                                    </div>
                                    <div className="flex gap-2">
                                        {appt.status === 'confirmed' && doesMeetingAllowStart(appt.start_time, appt.end_time) && (
                                            <Button onClick={() => handleStartMeeting(appt.id)}>Start</Button>
                                        )}
                                        {appt.status !== 'cancelled' && (
                                            <Button variant="outline" onClick={() => setCancelId(appt.id)}>Cancel</Button>
                                        )}
                                    </div>
                                </CardContent>
                            </Card>
                        ))
                    )}
                </div>

                <Dialog open={!!cancelId} onOpenChange={(open) => !open && setCancelId(null)}>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Cancel Appointment</DialogTitle>
                            <DialogDescription>
                                Are you sure you want to cancel this appointment? This action cannot be undone.
                            </DialogDescription>
                        </DialogHeader>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setCancelId(null)}>Keep Appointment</Button>
                            <Button variant="destructive" onClick={confirmCancel}>Yes, Cancel</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </div>
    );
}
