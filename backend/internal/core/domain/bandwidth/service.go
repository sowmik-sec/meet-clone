package bandwidth

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ReportBandwidth updates or creates a bandwidth record for a user in a room
func (s *Service) ReportBandwidth(ctx context.Context, userID, roomID, sessionID string, bytesSent, bytesReceived, packetsSent, packetsLost, duration int64) error {
	// Try to find existing record for this user and room
	// We group by room participation - if a user rejoins a room, we might want to sum it up or create a new one
	// For simplicity, we'll try to update the record for the current session/room
	record, err := s.repo.FindBySession(ctx, userID, roomID, sessionID)
	if err != nil {
		// If not found (or other error), we'll create a new one if it's really not found
		// But FindByUserAndRoom should return nil bandwidth if not found, not necessarily an error depending on impl
		// Let's assume repo implementation returns (nil, nil) if not found
	}

	if record == nil {
		// Create new record
		newRecord := &BandwidthRecord{
			ID:            uuid.New().String(),
			UserID:        userID,
			RoomID:        roomID,
			SessionID:     sessionID,
			BytesSent:     bytesSent,
			BytesReceived: bytesReceived,
			PacketsSent:   packetsSent,
			PacketsLost:   packetsLost,
			Duration:      duration,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err = s.repo.Create(ctx, newRecord)
		if err == nil {
			log.Printf("[BandwidthService] Created new record: UserID=%s, RoomID=%s, BytesSent=%d", userID, roomID, bytesSent)
		}
		return err
	}

	// Update existing record
	// Note: We're replacing the values because the client sends cumulative stats for the session
	// If the client sent deltas, we would add them.
	// Assuming client sends cumulative stats from WebRTC API
	record.BytesSent = bytesSent
	record.BytesReceived = bytesReceived
	record.PacketsSent = packetsSent
	record.PacketsLost = packetsLost
	record.Duration = duration
	record.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, record)
	if err == nil {
		log.Printf("[BandwidthService] Updated record: UserID=%s, RoomID=%s, BytesSent=%d", userID, roomID, bytesSent)
	}
	return err
}

func (s *Service) GetUserStats(ctx context.Context, userID string) (*UserBandwidthStats, error) {
	return s.repo.GetUserStats(ctx, userID)
}

func (s *Service) GetUserHistory(ctx context.Context, userID string, limit int) ([]*BandwidthRecord, error) {
	return s.repo.GetUserHistory(ctx, userID, limit)
}
