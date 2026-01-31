package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CallsService struct {
	accountID string
	apiToken  string
	appID     string
	baseURL   string
	client    *http.Client
}

func NewCallsService(accountID, apiToken, appID string) *CallsService {
	return &CallsService{
		accountID: accountID,
		apiToken:  apiToken,
		appID:     appID,
		baseURL:   "https://api.cloudflare.com/client/v4",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CreateSessionRequest struct {
	SessionDescription string `json:"sessionDescription,omitempty"`
}

type CreateSessionResponse struct {
	SessionID          string  `json:"sessionId"`
	SessionDescription string  `json:"sessionDescription"`
	Tracks             []Track `json:"tracks"`
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

// CreateSession creates a new Cloudflare Calls session
func (s *CallsService) CreateSession(roomID string) (*CreateSessionResponse, error) {
	// Validate that credentials are configured
	if s.accountID == "" || s.apiToken == "" || s.appID == "" {
		return nil, fmt.Errorf("cloudflare credentials not configured: accountID=%s, apiToken=%s, appID=%s",
			maskSecret(s.accountID), maskSecret(s.apiToken), s.appID)
	}

	// Using Cloudflare API v4 endpoint
	url := fmt.Sprintf("%s/accounts/%s/calls/apps/%s/sessions/new", s.baseURL, s.accountID, s.appID)

	// Create an empty JSON object for the request body
	reqBody := []byte("{}")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cloudflare API error (status %d): %s, body: %s", resp.StatusCode, resp.Status, string(body))
	}

	var result CreateSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	return &result, nil
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

// GenerateToken generates a token for joining a session
func (s *CallsService) GenerateToken(sessionID string) (*GenerateTokenResponse, error) {
	// Using Cloudflare API v4 endpoint
	url := fmt.Sprintf("%s/accounts/%s/calls/apps/%s/sessions/%s/tokens/new", s.baseURL, s.accountID, s.appID, sessionID)

	tokenReq := GenerateTokenRequest{
		SessionID: sessionID,
		TTL:       3600, // 1 hour
	}

	jsonData, err := json.Marshal(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cloudflare API error: %s, body: %s", resp.Status, string(body))
	}

	var result GenerateTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DeleteSession deletes a Cloudflare Calls session
func (s *CallsService) DeleteSession(sessionID string) error {
	url := fmt.Sprintf("%s/accounts/%s/calls/apps/%s/sessions/%s", s.baseURL, s.accountID, s.appID, sessionID)

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
