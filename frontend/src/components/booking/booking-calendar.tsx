import { Calendar } from "@/components/ui/calendar";

interface BookingCalendarProps {
    selectedDate: Date | undefined;
    onSelectDate: (date: Date | undefined) => void;
}

export function BookingCalendar({ selectedDate, onSelectDate }: BookingCalendarProps) {
    return (
        <div className="flex justify-center h-full items-start pt-4">
            <Calendar
                mode="single"
                selected={selectedDate}
                onSelect={onSelectDate}
                disabled={(date) => {
                    // Disable past dates
                    const today = new Date();
                    today.setHours(0, 0, 0, 0);
                    return date < today;
                }}
                className="rounded-md border shadow-sm p-4 bg-white"
                classNames={{
                    day_selected: "bg-black text-white hover:bg-black/90 focus:bg-black/90",
                    day_today: "bg-gray-100 text-gray-900 font-bold",
                }}
            />
        </div>
    );
}
