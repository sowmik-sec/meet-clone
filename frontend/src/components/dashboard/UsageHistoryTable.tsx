import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { billingApi } from '@/lib/api/billing';
import { UserBillingPeriod } from '@/types/billing';
import { Loader2 } from 'lucide-react';

export default function UsageHistoryTable() {
    const [history, setHistory] = useState<UserBillingPeriod[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const data = await billingApi.getHistory();
                setHistory(data || []);
            } catch (error) {
                console.error('Failed to fetch history:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchHistory();
    }, []);

    if (loading) {
        return (
            <Card className="mt-8">
                <CardContent className="flex justify-center items-center h-48">
                    <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
                </CardContent>
            </Card>
        );
    }

    if (history.length === 0) {
        return null;
    }

    return (
        <Card className="mt-8">
            <CardHeader>
                <CardTitle>Usage History</CardTitle>
                <CardDescription>
                    Monthly breakdown of your participant minutes usage
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div className="overflow-x-auto">
                    <table className="w-full text-sm text-left">
                        <thead className="bg-gray-50 text-gray-700 uppercase">
                            <tr>
                                <th className="px-6 py-3">Period</th>
                                <th className="px-6 py-3">Meetings</th>
                                <th className="px-6 py-3">Hosted</th>
                                <th className="px-6 py-3">Participant Minutes</th>
                                <th className="px-6 py-3">Overage</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                            {history.map((record) => (
                                <tr key={record.id} className="bg-white hover:bg-gray-50">
                                    <td className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">
                                        {new Date(record.billing_period + '-01').toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
                                    </td>
                                    <td className="px-6 py-4">{record.total_meetings}</td>
                                    <td className="px-6 py-4">{record.hosted_meetings}</td>
                                    <td className="px-6 py-4">{record.total_participant_minutes.toLocaleString()}</td>
                                    <td className="px-6 py-4">
                                        {record.overage_minutes > 0 ? (
                                            <span className="text-red-500">{record.overage_minutes.toLocaleString()}</span>
                                        ) : (
                                            <span className="text-green-500">0</span>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </CardContent>
        </Card>
    );
}
