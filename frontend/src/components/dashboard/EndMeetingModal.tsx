'use client';

import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

interface EndMeetingModalProps {
    isOpen: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void;
}

export function EndMeetingModal({ isOpen, onOpenChange, onConfirm }: EndMeetingModalProps) {
    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>End Meeting</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to end this meeting? This action cannot be undone and will remove all participants from the session.
                    </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>
                        Cancel
                    </Button>
                    <Button variant="destructive" onClick={onConfirm}>
                        End Meeting
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
