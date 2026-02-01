import { useEffect, useState } from 'react';
import { bandwidthApi, UserBandwidthStats } from '@/lib/api/bandwidth';

export default function BandwidthStatsCard() {
    const [stats, setStats] = useState<UserBandwidthStats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const data = await bandwidthApi.getStats();
                setStats(data);
            } catch (error) {
                console.error('Failed to fetch bandwidth stats:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    const formatBytes = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatDuration = (seconds: number) => {
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        return `${h}h ${m}m ${s}s`;
    };

    if (loading) {
        return (
            <div className="bg-gray-800 rounded-lg p-6 animate-pulse">
                <div className="h-4 bg-gray-700 rounded w-1/4 mb-4"></div>
                <div className="grid grid-cols-2 gap-4">
                    <div className="h-20 bg-gray-700 rounded"></div>
                    <div className="h-20 bg-gray-700 rounded"></div>
                </div>
            </div>
        );
    }

    if (!stats) return null;

    return (
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700 shadow-lg">
            <h3 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                Bandwidth Usage
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="bg-gray-700/50 rounded-lg p-4">
                    <p className="text-gray-400 text-sm mb-1">Data Sent</p>
                    <p className="text-2xl font-bold text-green-400">{formatBytes(stats.total_bytes_sent)}</p>
                </div>

                <div className="bg-gray-700/50 rounded-lg p-4">
                    <p className="text-gray-400 text-sm mb-1">Data Received</p>
                    <p className="text-2xl font-bold text-blue-400">{formatBytes(stats.total_bytes_received)}</p>
                </div>

                <div className="bg-gray-700/50 rounded-lg p-4">
                    <p className="text-gray-400 text-sm mb-1">Total Meeting Time</p>
                    <p className="text-2xl font-bold text-white">{formatDuration(stats.total_duration)}</p>
                </div>

                <div className="bg-gray-700/50 rounded-lg p-4">
                    <p className="text-gray-400 text-sm mb-1">Meetings Joined</p>
                    <p className="text-2xl font-bold text-purple-400">{stats.meeting_count}</p>
                </div>
            </div>
        </div>
    );
}
