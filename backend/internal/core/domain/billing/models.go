package billing

import "time"

// MeetingSession tracks an individual user's participation in a meeting
type MeetingSession struct {
	ID                  string     `bson:"_id" json:"id"`
	UserID              string     `bson:"user_id" json:"user_id"`
	RoomID              string     `bson:"room_id" json:"room_id"`
	RoomCreatorID       string     `bson:"room_creator_id" json:"room_creator_id"` // Tracks who pays for this session
	CloudflareSessionID string     `bson:"cloudflare_session_id" json:"cloudflare_session_id"`
	JoinedAt            time.Time  `bson:"joined_at" json:"joined_at"`
	LeftAt              *time.Time `bson:"left_at,omitempty" json:"left_at,omitempty"`
	ParticipantMinutes  int64      `bson:"participant_minutes" json:"participant_minutes"`
	IsHost              bool       `bson:"is_host" json:"is_host"`
	BillingPeriod       string     `bson:"billing_period" json:"billing_period"` // Format: "YYYY-MM"
}

// UserBillingPeriod represents aggregated monthly usage for a user
type UserBillingPeriod struct {
	ID                      string    `bson:"_id" json:"id"`
	UserID                  string    `bson:"user_id" json:"user_id"`
	BillingPeriod           string    `bson:"billing_period" json:"billing_period"` // Format: "YYYY-MM"
	TotalParticipantMinutes int64     `bson:"total_participant_minutes" json:"total_participant_minutes"`
	TotalMeetings           int       `bson:"total_meetings" json:"total_meetings"`
	HostedMeetings          int       `bson:"hosted_meetings" json:"hosted_meetings"`
	IncludedMinutes         int64     `bson:"included_minutes" json:"included_minutes"`
	OverageMinutes          int64     `bson:"overage_minutes" json:"overage_minutes"`
	CreatedAt               time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt               time.Time `bson:"updated_at" json:"updated_at"`
}

// CloudflareUsageSync tracks the state of usage synchronization with Cloudflare
type CloudflareUsageSync struct {
	ID           string    `bson:"_id" json:"id"`
	LastSyncAt   time.Time `bson:"last_sync_at" json:"last_sync_at"`
	SyncedUpTo   time.Time `bson:"synced_up_to" json:"synced_up_to"`
	Status       string    `bson:"status" json:"status"` // "success", "failed", "pending"
	ErrorMessage string    `bson:"error_message,omitempty" json:"error_message,omitempty"`
}
