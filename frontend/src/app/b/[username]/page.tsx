"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { Clock, ChevronRight, Video } from "lucide-react";

import { eventTypesApi } from "@/lib/api/event-types";
import { EventType } from "@/types/event-type";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Skeleton } from "../../../components/ui/skeleton";

export default function BookingProfilePage() {
    const params = useParams();
    const username = params.username as string; // This is actually userId currently
    const [eventTypes, setEventTypes] = useState<EventType[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(false);

    useEffect(() => {
        const fetchEventTypes = async () => {
            try {
                if (!username) return;
                const data = await eventTypesApi.listPublic(username);
                setEventTypes(data.filter(et => et.is_active));
            } catch (err) {
                console.error("Failed to load event types", err);
                setError(true);
            } finally {
                setLoading(false);
            }
        };

        fetchEventTypes();
    }, [username]);

    if (loading) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
                <div className="max-w-xl w-full space-y-4">
                    <div className="flex justify-center mb-8">
                        <Skeleton className="h-20 w-20 rounded-full" />
                    </div>
                    <Skeleton className="h-32 w-full rounded-lg" />
                    <Skeleton className="h-32 w-full rounded-lg" />
                </div>
            </div>
        );
    }

    if (error || eventTypes.length === 0) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
                <Card className="max-w-md w-full text-center p-8">
                    <CardTitle className="text-xl mb-2">No event types found</CardTitle>
                    <CardDescription>
                        This user hasn't set up any public event types yet.
                    </CardDescription>
                </Card>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">
            <header className="bg-white border-b sticky top-0 z-10">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-2 font-bold text-xl text-blue-600">
                        <Video className="w-6 h-6" />
                        <span>Meet Clone</span>
                    </div>
                    {/* Optional: Add Login link here later if needed */}
                </div>
            </header>
            <div className="flex-1 flex flex-col items-center py-12 px-4 sm:px-6 lg:px-8">
                <div className="w-full max-w-2xl space-y-8">
                    <div className="text-center">
                        <Avatar className="h-24 w-24 mx-auto border-4 border-white shadow-sm">
                            <AvatarImage src={`https://ui-avatars.com/api/?name=User&background=random`} />
                            <AvatarFallback>U</AvatarFallback>
                        </Avatar>
                        <h1 className="mt-4 text-2xl font-bold text-gray-900">Book a Meeting</h1>
                        <p className="mt-2 text-gray-600">
                            Select an event type to schedule a time.
                        </p>
                    </div>

                    <div className="grid gap-4">
                        {eventTypes.map((eventType) => (
                            <Link
                                key={eventType.id}
                                href={`/b/${username}/${eventType.slug}`}
                                className="block group"
                            >
                                <Card className="transition-all duration-200 hover:shadow-md hover:border-blue-500/50 cursor-pointer">
                                    <CardContent className="p-6 flex items-center justify-between">
                                        <div className="space-y-1">
                                            <div className="flex items-center gap-2">
                                                <div
                                                    className="w-3 h-3 rounded-full"
                                                    style={{ backgroundColor: eventType.color || '#3b82f6' }}
                                                />
                                                <h3 className="font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">
                                                    {eventType.title}
                                                </h3>
                                            </div>
                                            <p className="text-sm text-gray-500 line-clamp-1">
                                                {eventType.description || "No description"}
                                            </p>
                                            <div className="flex items-center text-sm text-gray-500 mt-2">
                                                <Clock className="w-4 h-4 mr-1.5" />
                                                {eventType.duration} mins
                                            </div>
                                        </div>
                                        <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-blue-500 transform group-hover:translate-x-1 transition-all" />
                                    </CardContent>
                                </Card>
                            </Link>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
