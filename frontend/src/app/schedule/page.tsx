"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { appointmentApi } from '@/lib/api/appointment';
import { Appointment } from '@/types/appointment';
import { EventType } from '@/types/event-type';
import { useAuth } from '@/hooks/useAuth';
import { Calendar, Clock, Video, Users, Plus, FileText } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';
import { eventTypesApi } from '@/lib/api/event-types';
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
    const [selectedAppointment, setSelectedAppointment] = useState<Appointment | null>(null);
    const [selectedEventType, setSelectedEventType] = useState<EventType | null>(null);
    const [activeTab, setActiveTab] = useState<'upcoming' | 'pending' | 'cancelled'>('upcoming');

    useEffect(() => {
        if (hasHydrated && !isAuthenticated) {
            router.push('/login');
        }
    }, [isAuthenticated, hasHydrated, router]);

    useEffect(() => {
        const fetchAppointments = async () => {
            setLoading(true);
            try {
                // Fetch all recent appointments
                // In a real app, you might want to filter by status on the backend
                // For now, we'll fetch all future/recent ones and filter client-side for the tabs
                const data = await appointmentApi.getAppointments({
                    start_time_after: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // Get from 24h ago
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

    const filteredAppointments = appointments.filter(appt => {
        if (activeTab === 'cancelled') return appt.status === 'cancelled';
        if (activeTab === 'pending') return appt.status === 'pending';
        // Upcoming defaults to confirmed
        return appt.status === 'confirmed';
    });


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

    const handleConfirm = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            await appointmentApi.confirmAppointment(id);
            setAppointments(appointments.map(a =>
                a.id === id ? { ...a, status: 'confirmed' } : a
            ));
            toast({
                title: "Appointment Confirmed",
                description: "You have approved this appointment request.",
            });
        } catch (error) {
            console.error('Failed to confirm appointment:', error);
            toast({
                title: "Error",
                description: "Failed to confirm appointment. Please try again.",
                variant: "destructive",
            });
        }
    };

    const confirmCancel = async () => {
        if (!cancelId) return;
        try {
            await appointmentApi.cancelAppointment(cancelId);
            setAppointments(appointments.map(a =>
                a.id === cancelId ? { ...a, status: 'cancelled' } : a
            ));
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

    const handleViewDetails = async (appt: Appointment) => {
        setSelectedAppointment(appt);
        setSelectedEventType(null); // Reset
        if (appt.event_type_id) {
            try {
                const et = await eventTypesApi.getById(appt.event_type_id);
                setSelectedEventType(et);
            } catch (error) {
                console.error("Failed to load event type details", error);
            }
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

                <div className="flex space-x-2 mb-6 border-b">
                    <button
                        className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'upcoming' ? 'border-primary text-primary' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
                        onClick={() => setActiveTab('upcoming')}
                    >
                        Upcoming
                    </button>
                    <button
                        className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'pending' ? 'border-primary text-primary' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
                        onClick={() => setActiveTab('pending')}
                    >
                        Pending
                    </button>
                    <button
                        className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'cancelled' ? 'border-primary text-primary' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
                        onClick={() => setActiveTab('cancelled')}
                    >
                        Cancelled
                    </button>
                </div>

                <div className="space-y-4">
                    {filteredAppointments.length === 0 ? (
                        <Card>
                            <CardContent className="p-8 text-center text-gray-500">
                                No {activeTab} appointments found.
                            </CardContent>
                        </Card>
                    ) : (
                        filteredAppointments.map((appt) => (
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
                                        <Button variant="outline" onClick={() => handleViewDetails(appt)}>View Details</Button>
                                        {appt.status === 'confirmed' && doesMeetingAllowStart(appt.start_time, appt.end_time) && (
                                            <Button onClick={() => handleStartMeeting(appt.id)}>Start</Button>
                                        )}
                                        {appt.status === 'pending' && (
                                            <Button
                                                className="bg-green-600 hover:bg-green-700"
                                                onClick={(e) => handleConfirm(appt.id, e)}
                                            >
                                                Approve
                                            </Button>
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

                <Dialog open={!!selectedAppointment} onOpenChange={(open) => !open && setSelectedAppointment(null)}>
                    <DialogContent className="sm:max-w-md">
                        <DialogHeader>
                            <DialogTitle>Appointment Details</DialogTitle>
                        </DialogHeader>
                        {selectedAppointment && (
                            <div className="space-y-4">
                                <div className="grid grid-cols-2 gap-4 text-sm">
                                    <div>
                                        <label className="text-gray-500 font-medium block">Guest</label>
                                        <div className="font-semibold">{selectedAppointment.title}</div>
                                    </div>
                                    <div>
                                        <label className="text-gray-500 font-medium block">Time</label>
                                        <div>{new Date(selectedAppointment.start_time).toLocaleDateString()}</div>
                                        <div className="text-gray-500">
                                            {new Date(selectedAppointment.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} -
                                            {new Date(selectedAppointment.end_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                        </div>
                                    </div>
                                    {selectedAppointment.description && (
                                        <div className="col-span-2">
                                            <label className="text-gray-500 font-medium block">Description</label>
                                            <div>{selectedAppointment.description}</div>
                                        </div>
                                    )}
                                </div>

                                {selectedAppointment.answers && Object.keys(selectedAppointment.answers).length > 0 && (
                                    <div className="border-t pt-4 mt-4">
                                        <h4 className="font-semibold mb-3 text-sm">Questions & Answers</h4>
                                        <div className="space-y-3">
                                            {Object.entries(selectedAppointment.answers).map(([qId, answer]) => {
                                                const question = selectedEventType?.questions?.find(q => q.id === qId);
                                                return (
                                                    <div key={qId} className="bg-gray-50 p-3 rounded-md">
                                                        <p className="text-xs text-gray-500 font-medium mb-1">
                                                            {question ? question.label : "Question"}
                                                        </p>
                                                        <p className="text-sm text-gray-900">{String(answer)}</p>
                                                    </div>
                                                );
                                            })}
                                        </div>
                                    </div>
                                )}
                            </div>
                        )}
                        <DialogFooter>
                            <Button onClick={() => setSelectedAppointment(null)}>Close</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </div >
    );
}
