package bandwidth

import (
	"context"
	"time"
)

// BandwidthRecord represents a bandwidth usage report for a user in a room
type BandwidthRecord struct {
	ID            string    `json:"id" bson:"_id"`
	UserID        string    `json:"user_id" bson:"user_id"`
	RoomID        string    `json:"room_id" bson:"room_id"`
	SessionID     string    `json:"session_id" bson:"session_id"`
	BytesSent     int64     `json:"bytes_sent" bson:"bytes_sent"`
	BytesReceived int64     `json:"bytes_received" bson:"bytes_received"`
	PacketsSent   int64     `json:"packets_sent" bson:"packets_sent"`
	PacketsLost   int64     `json:"packets_lost" bson:"packets_lost"`
	Duration      int64     `json:"duration" bson:"duration"` // duration in seconds
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

// UserBandwidthStats represents aggregated bandwidth statistics for a user
type UserBandwidthStats struct {
	UserID             string `json:"user_id" bson:"_id"`
	TotalBytesSent     int64  `json:"total_bytes_sent" bson:"total_bytes_sent"`
	TotalBytesReceived int64  `json:"total_bytes_received" bson:"total_bytes_received"`
	TotalDuration      int64  `json:"total_duration" bson:"total_duration"`
	MeetingCount       int    `json:"meeting_count" bson:"meeting_count"`
}

// Repository interface defines methods for bandwidth data access
type Repository interface {
	// Create inserts a new bandwidth record
	Create(ctx context.Context, record *BandwidthRecord) error

	// Update updates an existing bandwidth record
	Update(ctx context.Context, record *BandwidthRecord) error

	// FindBySession finds a bandwidth record for a specific user, room and session
	FindBySession(ctx context.Context, userID, roomID, sessionID string) (*BandwidthRecord, error)

	// GetUserStats calculates aggregated statistics for a user
	GetUserStats(ctx context.Context, userID string) (*UserBandwidthStats, error)

	// GetUserHistory retrieves bandwidth usage history for a user
	GetUserHistory(ctx context.Context, userID string, limit int) ([]*BandwidthRecord, error)
}
