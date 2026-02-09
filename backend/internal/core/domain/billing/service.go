package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/meet-clone/backend/internal/adapters/output/cloudflare"
)

type Service struct {
	repo Repository
	cf   *cloudflare.Client
}

func NewService(repo Repository, cf *cloudflare.Client) *Service {
	return &Service{
		repo: repo,
		cf:   cf,
	}
}

// StartSession creates a new meeting session record when a user joins
func (s *Service) StartSession(ctx context.Context, userID, roomID, cfSessionID, roomCreatorID string, isHost bool) error {
	now := time.Now()
	period := now.Format("2006-01") // YYYY-MM format

	session := &MeetingSession{
		ID:                  fmt.Sprintf("%s-%s-%d", userID, roomID, now.Unix()), // Simple ID generation
		UserID:              userID,
		RoomID:              roomID,
		RoomCreatorID:       roomCreatorID,
		CloudflareSessionID: cfSessionID,
		JoinedAt:            now,
		IsHost:              isHost,
		BillingPeriod:       period,
	}

	return s.repo.CreateSession(ctx, session)
}

// EndSession ends a meeting session and updates usage stats
func (s *Service) EndSession(ctx context.Context, userID, roomID string) error {
	now := time.Now()

	// comprehensive update at repository level handles duration calculation
	// 1. Find active session
	session, err := s.repo.FindActiveSession(ctx, userID, roomID)
	if err != nil {
		return fmt.Errorf("failed to find active session: %w", err)
	}
	if session == nil {
		return nil // Session already ended or not found
	}

	// Calculate duration
	duration := now.Sub(session.JoinedAt)
	minutes := int64(duration.Minutes())
	if minutes < 1 {
		minutes = 1 // Minimum 1 minute charge
	}

	// 2. End session in DB
	if err := s.repo.EndSession(ctx, userID, roomID, now, minutes); err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}

	// 3. Update monthly aggregation
	period := session.JoinedAt.Format("2006-01")
	return s.RecalculatePeriod(ctx, session.RoomCreatorID, period)
}

// GetCurrentPeriodUsage retrieves usage for the current month
func (s *Service) GetCurrentPeriodUsage(ctx context.Context, userID string) (*UserBillingPeriod, error) {
	period := time.Now().Format("2006-01")
	return s.repo.GetOrCreateBillingPeriod(ctx, userID, period)
}

// RecalculatePeriod recalculates totals for a billing period based on sessions
func (s *Service) RecalculatePeriod(ctx context.Context, userID, period string) error {
	// Calculate total minutes from sessions
	totalMinutes, err := s.repo.AggregateUserMinutes(ctx, userID, period)
	if err != nil {
		return fmt.Errorf("failed to aggregate minutes: %w", err)
	}

	// Calculate meeting counts
	totalMeetings, hostedMeetings, err := s.repo.AggregateMeetingCounts(ctx, userID, period)
	if err != nil {
		return fmt.Errorf("failed to aggregate meeting counts: %w", err)
	}

	// Get or create period record
	billingPeriod, err := s.repo.GetOrCreateBillingPeriod(ctx, userID, period)
	if err != nil {
		return fmt.Errorf("failed to get billing period: %w", err)
	}

	// Update totals
	billingPeriod.TotalParticipantMinutes = totalMinutes
	billingPeriod.TotalMeetings = totalMeetings
	billingPeriod.HostedMeetings = hostedMeetings
	billingPeriod.UpdatedAt = time.Now()

	// Calculate overage (placeholder logic for now)
	if billingPeriod.IncludedMinutes > 0 && totalMinutes > billingPeriod.IncludedMinutes {
		billingPeriod.OverageMinutes = totalMinutes - billingPeriod.IncludedMinutes
	} else {
		billingPeriod.OverageMinutes = 0
	}

	return s.repo.UpdateBillingPeriod(ctx, billingPeriod)
}

// GetBillingPeriod retrieves usage for a specific period
func (s *Service) GetBillingPeriod(ctx context.Context, userID, period string) (*UserBillingPeriod, error) {
	return s.repo.GetBillingPeriod(ctx, userID, period)
}

// GetUserHistory retrieves billing history for a user
func (s *Service) GetUserHistory(ctx context.Context, userID string, limit int) ([]*UserBillingPeriod, error) {
	return s.repo.GetUserBillingHistory(ctx, userID, limit)
}

// SyncCloudflareUsage fetches usage from Cloudflare and stores it for verification
func (s *Service) SyncCloudflareUsage(ctx context.Context) error {
	// Fetch last sync time
	lastSync, err := s.repo.GetLastSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last sync: %w", err)
	}

	start := time.Now().Add(-24 * time.Hour) // Default to last 24h
	if lastSync != nil {
		start = lastSync.LastSyncAt
	}
	end := time.Now()

	requests, err := s.cf.FetchAccountAnalytics(ctx, start, end)
	if err != nil {
		return fmt.Errorf("failed to fetch cloudflare usage: %w", err)
	}

	// Store sync result
	record := &CloudflareUsageSync{
		LastSyncAt:   end,
		SyncedUpTo:   end,
		Status:       "success",
		ErrorMessage: fmt.Sprintf("Synced %d requests", requests),
	}

	return s.repo.SaveSyncStatus(ctx, record)
}
