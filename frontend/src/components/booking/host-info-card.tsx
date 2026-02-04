import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Clock, Globe } from "lucide-react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

interface HostInfoProps {
    hostId: string;
    hostName?: string;
    duration?: number;
    description?: string;
}

export function HostInfoCard({ hostId, hostName, duration = 60, description }: HostInfoProps) {
    return (
        <div className="border-r h-full p-6 flex flex-col gap-6">
            <div className="flex flex-col gap-4">
                <Avatar className="h-16 w-16">
                    <AvatarImage src={`https://avatar.vercel.sh/${hostId}`} />
                    <AvatarFallback>{hostId.substring(0, 2).toUpperCase()}</AvatarFallback>
                </Avatar>
                <div>
                    <h2 className="text-muted-foreground text-sm font-medium uppercase tracking-wide">Host</h2>
                    <h1 className="text-2xl font-bold mt-1 text-foreground">{hostName || hostId}</h1>
                </div>
            </div>

            <div className="space-y-4">
                <div className="flex items-center gap-3 text-muted-foreground">
                    <Clock className="w-4 h-4" />
                    <span className="font-medium">{duration} min</span>
                </div>
                <div className="flex items-center gap-3 text-muted-foreground">
                    <Globe className="w-4 h-4" />
                    <span className="font-medium">Video Meeting</span>
                </div>
            </div>

            {description && (
                <div className="text-muted-foreground text-sm leading-relaxed">
                    {description}
                </div>
            )}
        </div>
    );
}
