import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { billingApi } from '@/lib/api/billing';
import { UserBillingPeriod } from '@/types/billing';
import { Loader2, Activity, Users, Clock } from 'lucide-react';

export default function UsageSummaryCard() {
    const [usage, setUsage] = useState<UserBillingPeriod | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchUsage = async () => {
            try {
                const data = await billingApi.getCurrentUsage();
                setUsage(data);
            } catch (error) {
                console.error('Failed to fetch usage:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchUsage();
    }, []);

    if (loading) {
        return (
            <Card>
                <CardContent className="flex justify-center items-center h-48">
                    <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
                </CardContent>
            </Card>
        );
    }

    if (!usage) {
        return null;
    }

    // Calculate percentages if needed, otherwise just display stats
    const formattedPeriod = new Date(usage.billing_period + '-01').toLocaleDateString('en-US', { month: 'long', year: 'numeric' });

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                        Total Participant Minutes
                    </CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{usage.total_participant_minutes.toLocaleString()}</div>
                    <p className="text-xs text-muted-foreground">
                        in {formattedPeriod}
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                        Meetings Hosted
                    </CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{usage.hosted_meetings}</div>
                    <p className="text-xs text-muted-foreground">
                        Total meetings: {usage.total_meetings}
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                        Overage Minutes
                    </CardTitle>
                    <Clock className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold text-red-500">{usage.overage_minutes.toLocaleString()}</div>
                    <p className="text-xs text-muted-foreground">
                        Included: {usage.included_minutes.toLocaleString()}
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                        Current Status
                    </CardTitle>
                    <div className="h-4 w-4 rounded-full bg-green-500" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">Active</div>
                    <p className="text-xs text-muted-foreground">
                        Billing in good standing
                    </p>
                </CardContent>
            </Card>
        </div>
    );
}
