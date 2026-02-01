import { useEffect, useRef } from 'react';
import { useRealtimeKitMeeting } from '@cloudflare/realtimekit-react';
import { bandwidthApi, BandwidthReport } from '@/lib/api/bandwidth';

interface BandwidthStats {
    bytesSent: number;
    bytesReceived: number;
    packetsSent: number;
    packetsLost: number;
}

export function useBandwidthStats(roomId: string) {
    const { meeting } = useRealtimeKitMeeting();
    const startTimeRef = useRef<number>(Date.now());
    const intervalRef = useRef<NodeJS.Timeout | null>(null);

    // Refs for aggregation
    const pcStatsMapRef = useRef<Map<RTCPeerConnection, BandwidthStats>>(new Map());
    const closedStatsSumRef = useRef<BandwidthStats>({
        bytesSent: 0,
        bytesReceived: 0,
        packetsSent: 0,
        packetsLost: 0,
    });

    // Track stats to return to component
    const currentStatsRef = useRef<BandwidthStats>({
        bytesSent: 0,
        bytesReceived: 0,
        packetsSent: 0,
        packetsLost: 0,
    });

    useEffect(() => {
        if (!meeting) return;

        const collectStats = async () => {
            // Get all captured active PCs
            let pcs: RTCPeerConnection[] = (window as any).__capturedPCs || [];

            // Fallback for single capture or meeting object if array is empty
            if (pcs.length === 0) {
                const singlePC = (window as any).__capturedPC || (meeting as any)?.peerConnection;
                if (singlePC) pcs.push(singlePC);
            }

            if (pcs.length === 0) {
                // console.warn('BandwidthStats: No PeerConnections found');
                return;
            }

            // Detect closed PCs (present in map but not in current pcs list)
            // We assume if it's gone from 'pcs', it's closed or removed
            const currentPCSet = new Set(pcs);
            for (const [pc, stats] of pcStatsMapRef.current.entries()) {
                if (!currentPCSet.has(pc) || pc.connectionState === 'closed') {
                    // This PC is effectively gone/closed. Add its last known stats to closedSum.
                    closedStatsSumRef.current.bytesSent += stats.bytesSent;
                    closedStatsSumRef.current.bytesReceived += stats.bytesReceived;
                    closedStatsSumRef.current.packetsSent += stats.packetsSent;
                    closedStatsSumRef.current.packetsLost += stats.packetsLost;

                    // Remove from map so we don't count it again or double-add on next loop
                    pcStatsMapRef.current.delete(pc);
                    console.log('BandwidthStats: PC removed/closed, archiving stats:', stats);
                }
            }

            // Collect active stats
            const allReportTypes = new Set<string>();

            for (const pc of pcs) {
                if (pc.connectionState === 'closed') continue;

                try {
                    const statsReport = await pc.getStats();
                    let pcBytesSent = 0;
                    let pcBytesReceived = 0;
                    let pcPacketsSent = 0;
                    let pcPacketsLost = 0;

                    statsReport.forEach((report) => {
                        allReportTypes.add(report.type);

                        // DEBUG: Log outbound-rtp details
                        if (report.type === 'outbound-rtp') {
                            console.log('BandwidthStats: outbound-rtp report:', {
                                isRemote: report.isRemote,
                                bytesSent: report.bytesSent,
                                packetsSent: report.packetsSent,
                                mediaType: (report as any).mediaType,
                                kind: (report as any).kind
                            });
                        }

                        if (report.type === 'outbound-rtp' && !report.isRemote) {
                            pcBytesSent += report.bytesSent || 0;
                            pcPacketsSent += report.packetsSent || 0;
                        }
                        if (report.type === 'inbound-rtp' && !report.isRemote) {
                            pcBytesReceived += report.bytesReceived || 0;
                            pcPacketsLost += report.packetsLost || 0;
                        }
                    });

                    // Get previous stats for this PC (if any)
                    const prevStats = pcStatsMapRef.current.get(pc);

                    // MONOTONIC: Keep maximum values to prevent drops when peers leave
                    const newStats: BandwidthStats = {
                        bytesSent: Math.max(pcBytesSent, prevStats?.bytesSent || 0),
                        bytesReceived: Math.max(pcBytesReceived, prevStats?.bytesReceived || 0),
                        packetsSent: Math.max(pcPacketsSent, prevStats?.packetsSent || 0),
                        packetsLost: Math.max(pcPacketsLost, prevStats?.packetsLost || 0)
                    };

                    pcStatsMapRef.current.set(pc, newStats);

                } catch (err) {
                    console.error('Error getting stats from a PC:', err);
                }
            }

            // Calculate Grand Total
            let totalBytesSent = closedStatsSumRef.current.bytesSent;
            let totalBytesReceived = closedStatsSumRef.current.bytesReceived;
            let totalPacketsSent = closedStatsSumRef.current.packetsSent;
            let totalPacketsLost = closedStatsSumRef.current.packetsLost;

            for (const s of pcStatsMapRef.current.values()) {
                totalBytesSent += s.bytesSent;
                totalBytesReceived += s.bytesReceived;
                totalPacketsSent += s.packetsSent;
                totalPacketsLost += s.packetsLost;
            }

            // DEBUG: If 0 bytes, log what types we DID find
            if (totalBytesSent === 0 && totalBytesReceived === 0) {
                console.log('BandwidthStats: 0 bytes found. Available Report Types:', Array.from(allReportTypes));
            }

            // Update current stats ref for return
            currentStatsRef.current = {
                bytesSent: totalBytesSent,
                bytesReceived: totalBytesReceived,
                packetsSent: totalPacketsSent,
                packetsLost: totalPacketsLost,
            };

            const duration = Math.floor((Date.now() - startTimeRef.current) / 1000);

            // Report to backend
            const report: BandwidthReport = {
                room_id: roomId,
                session_id: (meeting as any).sessionId || 'unknown',
                bytes_sent: totalBytesSent,
                bytes_received: totalBytesReceived,
                packets_sent: totalPacketsSent,
                packets_lost: totalPacketsLost,
                duration: duration,
            };

            console.log('BandwidthStats: Reporting', report);
            try {
                await bandwidthApi.reportBandwidth(report);
            } catch (error) {
                console.error('Failed to report bandwidth stats:', error);
            }
        };

        // Start collection interval
        intervalRef.current = setInterval(collectStats, 10000); // Every 10 seconds

        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
            // Report one last time on unmount
            collectStats();
        };
    }, [meeting, roomId]);

    return currentStatsRef.current;
}
