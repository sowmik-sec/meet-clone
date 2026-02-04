import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { format } from "date-fns";
import { Loader2 } from "lucide-react";

interface TimeSlotGridProps {
    date: Date | undefined;
    slots: string[];
    selectedSlot: string | null;
    onSelectSlot: (slot: string) => void;
    loading?: boolean;
    isSlotAvailable: (time: string, date: Date) => boolean;
}

export function TimeSlotGrid({
    date,
    slots,
    selectedSlot,
    onSelectSlot,
    loading,
    isSlotAvailable
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
                        return (
                            <Button
                                key={time}
                                variant={selectedSlot === time ? "default" : "outline"}
                                className={`w-full justify-center transition-all ${selectedSlot === time
                                        ? "ring-2 ring-primary ring-offset-2"
                                        : ""
                                    }`}
                                onClick={() => available && onSelectSlot(time)}
                                disabled={!available}
                            >
                                {time}
                            </Button>
                        );
                    })}
                </div>
            </ScrollArea>
        </div>
    );
}
