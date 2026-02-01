'use client';

import { useState } from 'react';
import { Check, Copy, Link as LinkIcon } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface MeetingCreatedModalProps {
    isOpen: boolean;
    onOpenChange: (open: boolean) => void;
    roomId: string;
    onJoinMeeting: () => void;
}

export function MeetingCreatedModal({
    isOpen,
    onOpenChange,
    roomId,
    onJoinMeeting,
}: MeetingCreatedModalProps) {
    const [copiedLink, setCopiedLink] = useState(false);
    const [copiedId, setCopiedId] = useState(false);

    const getMeetingLink = () => {
        if (typeof window !== 'undefined') {
            return `${window.location.origin}/room/${roomId}`;
        }
        return '';
    };

    const copyToClipboard = async (text: string, type: 'link' | 'id') => {
        try {
            await navigator.clipboard.writeText(text);
            if (type === 'link') {
                setCopiedLink(true);
                setTimeout(() => setCopiedLink(false), 2000);
            } else {
                setCopiedId(true);
                setTimeout(() => setCopiedId(false), 2000);
            }
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>Meeting Created!</DialogTitle>
                    <DialogDescription>
                        Your meeting is ready. Share the link or ID with others you want to invite.
                    </DialogDescription>
                </DialogHeader>

                <div className="grid gap-6 py-4">
                    <div className="grid gap-2">
                        <Label htmlFor="meeting-link">Meeting Link</Label>
                        <div className="flex items-center gap-2">
                            <div className="relative flex-1">
                                <LinkIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
                                <Input
                                    id="meeting-link"
                                    value={getMeetingLink()}
                                    readOnly
                                    className="pl-9"
                                />
                            </div>
                            <Button
                                type="button"
                                size="icon"
                                variant="outline"
                                onClick={() => copyToClipboard(getMeetingLink(), 'link')}
                                title="Copy link"
                            >
                                {copiedLink ? (
                                    <Check className="h-4 w-4 text-green-500" />
                                ) : (
                                    <Copy className="h-4 w-4" />
                                )}
                                <span className="sr-only">Copy link</span>
                            </Button>
                        </div>
                    </div>

                    <div className="grid gap-2">
                        <Label htmlFor="meeting-id">Meeting ID</Label>
                        <div className="flex items-center gap-2">
                            <Input
                                id="meeting-id"
                                value={roomId}
                                readOnly
                                className="font-mono"
                            />
                            <Button
                                type="button"
                                size="icon"
                                variant="outline"
                                onClick={() => copyToClipboard(roomId, 'id')}
                                title="Copy ID"
                            >
                                {copiedId ? (
                                    <Check className="h-4 w-4 text-green-500" />
                                ) : (
                                    <Copy className="h-4 w-4" />
                                )}
                                <span className="sr-only">Copy ID</span>
                            </Button>
                        </div>
                    </div>
                </div>

                <DialogFooter className="sm:justify-between sm:space-x-2">
                    <Button
                        type="button"
                        variant="secondary"
                        onClick={() => onOpenChange(false)}
                    >
                        Close
                    </Button>
                    <Button type="button" onClick={onJoinMeeting} className="w-full sm:w-auto">
                        Join Now
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
