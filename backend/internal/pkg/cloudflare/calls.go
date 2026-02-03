package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type CallsService struct {
	accountID string
	apiToken  string
	appID     string
	appSecret string
	baseURL   string
	client    *http.Client
}

func NewCallsService(accountID, apiToken, appID, appSecret string) *CallsService {
	return &CallsService{
		accountID: accountID,
		apiToken:  apiToken,
		appID:     appID,
		appSecret: appSecret,
		baseURL:   "https://api.cloudflare.com/client/v4",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CreateSessionRequest struct {
	Title string `json:"title,omitempty"`
}

type CreateSessionResponse struct {
	SessionID          string  `json:"sessionId"`
	MeetingID          string  `json:"meetingId"`
	SessionDescription string  `json:"sessionDescription,omitempty"`
	Tracks             []Track `json:"tracks,omitempty"`
}

type Track struct {
	TrackName string `json:"trackName"`
	Location  string `json:"location"`
	Mid       string `json:"mid"`
}

type GenerateTokenRequest struct {
	SessionID string   `json:"sessionId"`
	Tracks    []string `json:"tracks,omitempty"`
	TTL       int      `json:"ttl,omitempty"` // in seconds
}

type GenerateTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// CreateSession creates a new Cloudflare RealtimeKit meeting and returns meeting ID
func (s *CallsService) CreateSession(roomID string) (*CreateSessionResponse, error) {
	// Validate that credentials are configured
	if s.accountID == "" || s.apiToken == "" || s.appID == "" {
		return nil, fmt.Errorf("cloudflare credentials not configured: accountID=%s, apiToken=%s, appID=%s",
			maskSecret(s.accountID), maskSecret(s.apiToken), s.appID)
	}

	// Step 1: Create a meeting using RealtimeKit API
	meetingURL := fmt.Sprintf("%s/accounts/%s/realtime/kit/%s/meetings", s.baseURL, s.accountID, s.appID)

	log.Printf("[Cloudflare] Creating meeting - URL: %s", meetingURL)
	log.Printf("[Cloudflare] AccountID: %s, AppID: %s", maskSecret(s.accountID), s.appID)

	meetingReq := map[string]interface{}{
		"title": fmt.Sprintf("Room %s", roomID),
	}

	// Log the request body for debugging
	reqJSON, _ := json.MarshalIndent(meetingReq, "", "  ")
	log.Printf("[Cloudflare] Meeting Request Body: %s", string(reqJSON))

	meetingBody, err := json.Marshal(meetingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meeting request: %w", err)
	}

	req, err := http.NewRequest("POST", meetingURL, bytes.NewBuffer(meetingBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create meeting request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[Cloudflare] Meeting request failed: %v", err)
		return nil, fmt.Errorf("failed to create meeting: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	log.Printf("[Cloudflare] Meeting Response Status: %d, Body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cloudflare API error (status %d): %s, body: %s", resp.StatusCode, resp.Status, string(body))
	}

	var meetingResp struct {
		Success bool `json:"success"`
		Data    struct {
			MeetingID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &meetingResp); err != nil {
		return nil, fmt.Errorf("failed to parse meeting response: %w, body: %s", err, string(body))
	}

	if !meetingResp.Success || meetingResp.Data.MeetingID == "" {
		return nil, fmt.Errorf("failed to get meeting ID from response: %s", string(body))
	}

	return &CreateSessionResponse{
		SessionID: meetingResp.Data.MeetingID,
		MeetingID: meetingResp.Data.MeetingID,
	}, nil
}

// maskSecret masks the secret for logging purposes
func maskSecret(secret string) string {
	if secret == "" {
		return "<empty>"
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// GenerateToken generates a participant token by calling Cloudflare's API
// isCreator determines if participant gets moderator preset (auto-admit) or participant preset (needs approval)
// isWebinar determines if webinar presets should be used instead of group call presets
func (s *CallsService) GenerateToken(sessionID, participantID string, isCreator bool, isWebinar bool) (*GenerateTokenResponse, error) {
	// Call Cloudflare API to create participant and get auth token
	participantURL := fmt.Sprintf("%s/accounts/%s/realtime/kit/%s/meetings/%s/participants",
		s.baseURL, s.accountID, s.appID, sessionID)

	log.Printf("[Cloudflare] Creating participant via API - URL: %s, ParticipantID: %s, IsCreator: %v, IsWebinar: %v", participantURL, participantID, isCreator, isWebinar)

	// Assign preset based on room type and creator status
	var presetName string
	if isWebinar {
		// Webinar mode: presenters can broadcast, viewers can only watch
		if isCreator {
			presetName = "webinar_presenter"
		} else {
			presetName = "webinar_viewer"
		}
	} else {
		// Meeting mode: all participants can share audio/video
		if isCreator {
			presetName = "group_call_host"
		} else {
			presetName = "group_call_participant"
		}
	}

	participantReq := map[string]interface{}{
		"custom_participant_id": participantID,
		"name":                  "Participant",
		"preset_name":           presetName,
	}

	jsonData, err := json.Marshal(participantReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[Cloudflare] Participant Request Body: %s", string(jsonData))

	req, err := http.NewRequest("POST", participantURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create participant: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	log.Printf("[Cloudflare] Participant Response Status: %d, Body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cloudflare API error: %s, body: %s", resp.Status, string(body))
	}

	var participantResp struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &participantResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !participantResp.Success || participantResp.Data.Token == "" {
		return nil, fmt.Errorf("failed to get auth token from response: %s", string(body))
	}

	log.Printf("[Cloudflare] Successfully generated auth token for participant %s", participantID)

	return &GenerateTokenResponse{
		Token:     participantResp.Data.Token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// DeleteSession deletes a Cloudflare RealtimeKit meeting
func (s *CallsService) DeleteSession(sessionID string) error {
	url := fmt.Sprintf("%s/accounts/%s/realtime/kit/%s/meetings/%s", s.baseURL, s.accountID, s.appID, sessionID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiToken))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare API error: %s, body: %s", resp.Status, string(body))
	}

	return nil
}
