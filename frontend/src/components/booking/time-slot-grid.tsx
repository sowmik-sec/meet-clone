import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { format } from "date-fns";
import { Loader2 } from "lucide-react";

export interface TimeSlotGridProps {
    date: Date | undefined;
    slots: string[];
    selectedSlot: string | null;
    onSelectSlot: (slot: string) => void;
    loading?: boolean;
    isSlotAvailable: (time: string, date: Date) => boolean;
    partialSlots?: Record<string, number>;
    maxAttendees?: number;
}

export function TimeSlotGrid({
    date,
    slots,
    selectedSlot,
    onSelectSlot,
    loading,
    isSlotAvailable,
    partialSlots = {},
    maxAttendees = 1
}: TimeSlotGridProps) {
    if (!date) {
        return (
            <div className="h-full flex items-center justify-center text-muted-foreground text-sm">
                Select a date to view available times
            </div>
        );
    }

    if (loading) {
        return (
            <div className="h-full flex items-center justify-center">
                <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (slots.length === 0) {
        return (
            <div className="h-full flex flex-col items-center justify-center text-muted-foreground gap-2">
                <span className="font-medium">No slots available</span>
                <span className="text-xs">Try selecting another date</span>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full">
            <div className="mb-4">
                <h3 className="font-medium text-foreground">
                    {format(date, "EEEE, MMMM d")}
                </h3>
            </div>
            <ScrollArea className="flex-1 -mr-2 pr-2">
                <div className="grid grid-cols-1 gap-2">
                    {slots.map((time) => {
                        const available = isSlotAvailable(time, date);
                        // Construct key for partial slots lookup (RFC3339 start|end)
                        // This logic must match backend key generation
                        // Backend: appt.BufferedStartTime.Format(time.RFC3339)|appt.BufferedEndTime.Format(time.RFC3339)
                        // But here we only have start time string "HH:MM".
                        // We need to construct the full ISO strings to match the key from backend.
                        // However, exact match might be tricky due to timezone/seconds.
                        // Ideally, we should receive partialSlots keyed by "HH:MM" or just return capacity info with slots from API.
                        // But since we generate slots locally in frontend, we don't have that direct mapping easily.

                        // Easier approach: Backend returned map key is absolute time.
                        // We can construct absolute start time here.
                        const dateStr = format(date, "yyyy-MM-dd");
                        const startTime = new Date(`${dateStr}T${time}:00`);
                        // We need to match the timezone used by backend or format it to RFC3339
                        // Backend uses RFC3339.
                        // Let's iterate partialSlots keys and find one that starts with our time?
                        // "2024-02-10T10:00:00Z|..."
                        // This is getting complicated due to timezone differences (local browser vs server).

                        // Alternative: Just match by HH:MM if we assume same timezone or convert both to local.
                        // Let's try to match by converting browser time to ISO string and seeing if it matches prefix?
                        // Or better: update backend key to be simpler? No, keep backend robust.

                        // Let's try to find a key that starts with our ISO timestamp.
                        // Note: toISOString() uses UTC. Backend key uses RFC3339 which is also typically UTC/offset.
                        const isoStart = startTime.toISOString().split('.')[0]; // Remove ms

                        let bookedCount = 0;
                        if (partialSlots) {
                            for (const [key, count] of Object.entries(partialSlots)) {
                                if (key.includes(isoStart)) { // Simple contains check
                                    bookedCount = count;
                                    break;
                                }
                            }
                        }



                        const remaining = maxAttendees - bookedCount;
                        const showSpotsLeft = available && maxAttendees > 1 && remaining > 0;

                        return (
                            <Button
                                key={time}
                                variant={selectedSlot === time ? "default" : "outline"}
                                className={`w-full justify-center transition-all flex flex-col items-center py-6 ${selectedSlot === time
                                    ? "ring-2 ring-primary ring-offset-2"
                                    : ""
                                    }`}
                                onClick={() => available && onSelectSlot(time)}
                                disabled={!available}
                            >
                                <span className="font-medium">{time}</span>
                                {showSpotsLeft && (
                                    <span className="text-[10px] font-normal opacity-80 mt-1">
                                        {remaining} spots left
                                    </span>
                                )}
                            </Button>
                        );
                    })}
                </div>
            </ScrollArea>
        </div>
    );
}
