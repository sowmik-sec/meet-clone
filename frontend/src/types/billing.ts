export interface UserBillingPeriod {
    id: string;
    user_id: string;
    billing_period: string;
    total_participant_minutes: number;
    total_meetings: number;
    hosted_meetings: number;
    included_minutes: number;
    overage_minutes: number;
    created_at: string;
    updated_at: string;
}

export interface UsageHistoryResponse {
    history: UserBillingPeriod[];
}
