"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { useToast } from '@/components/ui/use-toast';
import { appointmentApi } from '@/lib/api/appointment';
import { Appointment } from '@/types/appointment';
import { Calendar, Clock, Video, AlertTriangle } from 'lucide-react';
import { format, differenceInMinutes } from 'date-fns';

export default function AppointmentJoinPage() {
    const params = useParams();
    const id = params.id as string;
    const router = useRouter();
    const { toast } = useToast();

    const [appointment, setAppointment] = useState<Appointment | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchAppointment = async () => {
            try {
                // We need a public endpoint to get appointment details by ID without auth if possible?
                // Or we require guest to look it up?
                // The current API might require auth for GetAppointment. 
                // However, for public join links, we usually need a token or allow public read by ID if it's a valid ID.
                // Let's try fetching. If it fails due to auth, we might need a public endpoint or token.
                // Actually, wait. The email link is just UUID. UUID is hard to guess, but technically public read access to all appointments by ID is risky.
                // secure way: /appointments/:id?token=...
                // But typically for "Join Link", if you have the ID you can view basic info.
                // Let's assume for MVP we allow fetching by ID or check if endpoint allows it.
                // If not, we might fail here.
                // Verify backend: GetAppointment checks ownership? 

                const data = await appointmentApi.getAppointment(id);
                setAppointment(data);
            } catch (err: any) {
                console.error("Failed to load appointment", err);
                setError("Appointment not found or you don't have permission to view it.");
            } finally {
                setLoading(false);
            }
        };

        if (id) fetchAppointment();
    }, [id]);

    const handleJoin = () => {
        if (!appointment?.room_id) {
            toast({
                title: "Meeting Not Started",
                description: "The host has not started the meeting yet. Please wait.",
                variant: "destructive"
            });
            return;
        }
        router.push(`/room/${appointment.room_id}`);
    };

    if (loading) return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="animate-pulse flex flex-col items-center">
                <div className="h-12 w-12 bg-gray-200 rounded-full mb-4"></div>
                <div className="h-4 w-32 bg-gray-200 rounded"></div>
            </div>
        </div>
    );

    if (error || !appointment) return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
            <Card className="max-w-md w-full text-center p-8">
                <div className="flex justify-center mb-4">
                    <AlertTriangle className="h-12 w-12 text-yellow-500" />
                </div>
                <h2 className="text-xl font-semibold text-gray-900 mb-2">Unavailable</h2>
                <p className="text-gray-500">{error || "Appointment not found."}</p>
                <Button onClick={() => router.push('/')} className="mt-6" variant="outline">Go Home</Button>
            </Card>
        </div>
    );

    const now = new Date();
    const start = new Date(appointment.start_time);
    const end = new Date(appointment.end_time);
    const minutesToStart = differenceInMinutes(start, now);
    const isEnded = now > end;
    const isStarted = appointment.status === 'confirmed' && !!appointment.room_id; // Or simply room_id exists?
    // Actually room_id is created when host starts it.

    // Allow join if within 15 mins of start or already started?
    const canJoin = appointment.room_id || (minutesToStart <= 15 && !isEnded);

    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
            <Card className="max-w-xl w-full shadow-lg">
                <CardHeader className="text-center border-b bg-white rounded-t-lg pb-8 pt-8">
                    <div className="mx-auto bg-blue-100 w-16 h-16 rounded-full flex items-center justify-center mb-4">
                        <Video className="w-8 h-8 text-blue-600" />
                    </div>
                    <CardTitle className="text-2xl font-bold">{appointment.title}</CardTitle>
                    <CardDescription className="text-lg mt-2">
                        Hosted by {appointment.host_id}
                        {/* ideally we fetch host name too */}
                    </CardDescription>
                </CardHeader>
                <CardContent className="p-8 space-y-6">
                    <div className="grid gap-4 bg-gray-50 p-6 rounded-lg border border-gray-100">
                        <div className="flex items-center gap-3">
                            <Calendar className="w-5 h-5 text-gray-500" />
                            <span className="font-medium text-gray-900">
                                {format(start, 'EEEE, MMMM d, yyyy')}
                            </span>
                        </div>
                        <div className="flex items-center gap-3">
                            <Clock className="w-5 h-5 text-gray-500" />
                            <span className="font-medium text-gray-900">
                                {format(start, 'HH:mm')} - {format(end, 'HH:mm')}
                            </span>
                            <span className="text-xs text-gray-400 bg-gray-200 px-2 py-0.5 rounded-full">
                                {Intl.DateTimeFormat().resolvedOptions().timeZone}
                            </span>
                        </div>
                    </div>

                    <div className="space-y-4">
                        {isEnded ? (
                            <div className="text-center p-4 bg-red-50 text-red-700 rounded-lg">
                                This meeting has ended.
                            </div>
                        ) : appointment.status === 'cancelled' ? (
                            <div className="text-center p-4 bg-red-50 text-red-700 rounded-lg">
                                This meeting has been cancelled.
                            </div>
                        ) : (
                            <>
                                {appointment.room_id ? (
                                    <Button className="w-full h-12 text-lg" onClick={handleJoin}>
                                        Join Meeting Now
                                    </Button>
                                ) : (
                                    <div className="text-center space-y-3">
                                        <div className="p-4 bg-yellow-50 text-yellow-800 rounded-lg text-sm">
                                            The host has not started the meeting yet.
                                            <br />
                                            The button will appear here once the meeting starts.
                                        </div>
                                        <Button disabled className="w-full h-12 text-lg">
                                            Waiting for Host...
                                        </Button>
                                    </div>
                                )}

                                <div className="text-center text-sm text-gray-500 mt-4">
                                    Need to reschedule? <a href={`/reschedule/${appointment.reschedule_token}`} className="text-blue-600 hover:underline">Click here</a>
                                </div>
                            </>
                        )}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
