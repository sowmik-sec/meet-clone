import { useEffect, useState } from 'react';
import { Activity } from 'lucide-react';
import { bandwidthApi, UserBandwidthStats } from '@/lib/api/bandwidth';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

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
            <Card className="bg-gray-800 border-gray-700 animate-pulse">
                <CardContent className="p-6">
                    <div className="h-4 bg-gray-700 rounded w-1/4 mb-4"></div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="h-20 bg-gray-700 rounded"></div>
                        <div className="h-20 bg-gray-700 rounded"></div>
                    </div>
                </CardContent>
            </Card>
        );
    }

    if (!stats) return null;

    return (
        <Card className="border-gray-700 bg-gray-800 shadow-lg">
            <CardHeader>
                <CardTitle className="flex items-center gap-2 text-xl font-bold text-white">
                    <Activity className="h-6 w-6 text-blue-400" />
                    Bandwidth Usage
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <Card className="bg-gray-700/50 border-none">
                        <CardContent className="p-4">
                            <p className="mb-1 text-sm text-gray-400">Data Sent</p>
                            <p className="text-2xl font-bold text-green-400">{formatBytes(stats.total_bytes_sent)}</p>
                        </CardContent>
                    </Card>

                    <Card className="bg-gray-700/50 border-none">
                        <CardContent className="p-4">
                            <p className="mb-1 text-sm text-gray-400">Data Received</p>
                            <p className="text-2xl font-bold text-blue-400">{formatBytes(stats.total_bytes_received)}</p>
                        </CardContent>
                    </Card>

                    <Card className="bg-gray-700/50 border-none">
                        <CardContent className="p-4">
                            <p className="mb-1 text-sm text-gray-400">Total Meeting Time</p>
                            <p className="text-2xl font-bold text-white">{formatDuration(stats.total_duration)}</p>
                        </CardContent>
                    </Card>

                    <Card className="bg-gray-700/50 border-none">
                        <CardContent className="p-4">
                            <p className="mb-1 text-sm text-gray-400">Meetings Joined</p>
                            <p className="text-2xl font-bold text-purple-400">{stats.meeting_count}</p>
                        </CardContent>
                    </Card>
                </div>
            </CardContent>
        </Card>
    );
}
