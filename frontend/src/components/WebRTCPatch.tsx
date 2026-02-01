'use client';

// This component solely exists to monkey-patch the RTCPeerConnection
// so we can capture the internal connection instance used by RealtimeKit/Cloudflare Calls.
// This is required because the SDK does not expose the active PeerConnection in its public API.

if (typeof window !== 'undefined') {
    // Only patch once
    if (!(window as any).__patchApplied) {
        console.log('[WebRTCPatch] Initializing RTCPeerConnection patch...');

        const originalRTCPeerConnection = window.RTCPeerConnection;

        // @ts-ignore
        window.RTCPeerConnection = function (...args: any[]) {
            console.log('[WebRTCPatch] Intercepted new RTCPeerConnection creation');
            // @ts-ignore
            const pc = new originalRTCPeerConnection(...args);

            // Capture the instance globally
            if (!(window as any).__capturedPCs) {
                (window as any).__capturedPCs = [];
            }
            (window as any).__capturedPCs.push(pc);
            (window as any).__capturedPC = pc; // Legacy support (last one)
            (window as any).__activePeerConnection = pc;

            console.log(`[WebRTCPatch] Captured PC. Total count: ${(window as any).__capturedPCs.length}`);

            // Cleanup on close
            const removePc = () => {
                const index = (window as any).__capturedPCs.indexOf(pc);
                if (index > -1) {
                    (window as any).__capturedPCs.splice(index, 1);
                    console.log(`[WebRTCPatch] PC Closed/Removed. Total count: ${(window as any).__capturedPCs.length}`);
                }
            };
            pc.addEventListener('close', removePc);

            // Add event listeners for debugging
            pc.addEventListener('connectionstatechange', () => {
                console.log('[WebRTCPatch] Connection State Change:', pc.connectionState);
            });

            return pc;
        };

        // Restore prototype to ensure instance checks pass
        // @ts-ignore
        window.RTCPeerConnection.prototype = originalRTCPeerConnection.prototype;
        // Copy static properties
        Object.keys(originalRTCPeerConnection).forEach(key => {
            // @ts-ignore
            window.RTCPeerConnection[key] = originalRTCPeerConnection[key];
        });

        (window as any).__patchApplied = true;
        console.log('[WebRTCPatch] Patch applied successfully');
    }
}

export function WebRTCPatch() {
    return null;
}
