package billing

import (
	"context"
	"time"
)

// Repository defines the interface for billing data persistence
type Repository interface {
	// Session tracking
	CreateSession(ctx context.Context, session *MeetingSession) error
	UpdateSession(ctx context.Context, session *MeetingSession) error
	FindActiveSession(ctx context.Context, userID, roomID string) (*MeetingSession, error)
	EndSession(ctx context.Context, userID, roomID string, leftAt time.Time, durationMinutes int64) error

	// Billing period aggregation
	GetOrCreateBillingPeriod(ctx context.Context, userID, period string) (*UserBillingPeriod, error)
	UpdateBillingPeriod(ctx context.Context, period *UserBillingPeriod) error
	GetUserBillingHistory(ctx context.Context, userID string, limit int) ([]*UserBillingPeriod, error)
	GetBillingPeriod(ctx context.Context, userID, period string) (*UserBillingPeriod, error)

	// Aggregation queries
	AggregateUserMinutes(ctx context.Context, userID, period string) (int64, error)
	AggregateMeetingCounts(ctx context.Context, userID, period string) (totalMeetings, hostedMeetings int, err error)
	GetAllUsersUsage(ctx context.Context, period string) ([]*UserBillingPeriod, error)

	// Sync tracking
	GetLastSync(ctx context.Context) (*CloudflareUsageSync, error)
	SaveSyncStatus(ctx context.Context, sync *CloudflareUsageSync) error
}
