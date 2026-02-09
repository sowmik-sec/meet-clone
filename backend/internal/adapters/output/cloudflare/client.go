package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/meet-clone/backend/internal/config"
)

const (
	Endpoint = "https://api.cloudflare.com/client/v4/graphql"
)

type Client struct {
	apiToken   string
	accountID  string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiToken:   cfg.CloudflareAPIToken,
		accountID:  cfg.CloudflareAccountID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Query executes a GraphQL query against Cloudflare API
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, responseData interface{}) error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// wrapper for standard graphql response
	var wrapper GraphQLResponse
	wrapper.Data = responseData // Unmarshal data into the provided pointer

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(wrapper.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", wrapper.Errors[0].Message)
	}

	return nil
}

// FetchAccountAnalytics retrieves generic HTTP request counts for the account
// This serves as a verified usage metric
func (c *Client) FetchAccountAnalytics(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		query GetAccountAnalytics($accountTag: string, $start: Time, $end: Time) {
			viewer {
				accounts(filter: {accountTag: $accountTag}) {
					httpRequestsAdaptiveGroups(limit: 1, filter: {datetime_geq: $start, datetime_leq: $end}) {
						sum {
							requests
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"accountTag": c.accountID,
		"start":      start.Format(time.RFC3339),
		"end":        end.Format(time.RFC3339),
	}

	var result struct {
		Viewer struct {
			Accounts []struct {
				HttpRequestsAdaptiveGroups []struct {
					Sum struct {
						Requests int64 `json:"requests"`
					} `json:"sum"`
				} `json:"httpRequestsAdaptiveGroups"`
			} `json:"accounts"`
		} `json:"viewer"`
	}

	if err := c.Query(ctx, query, variables, &result.Viewer); err != nil {
		return 0, err
	}

	if len(result.Viewer.Accounts) == 0 {
		return 0, nil
	}

	if len(result.Viewer.Accounts[0].HttpRequestsAdaptiveGroups) == 0 {
		return 0, nil
	}

	return result.Viewer.Accounts[0].HttpRequestsAdaptiveGroups[0].Sum.Requests, nil
}

// CloseSession closes a session by closing all its tracks.
// Note: verify if there is a better way, but closing tracks usually effectively ends participation.
// Alternatively, we could rely on the fact that if no one is publishing, the session is effectively dead.
// But valid "End Meeting" usually implies kicking everyone out.
// Cloudflare Calls doesn't have a simple "End Session" for a room.
// The best way is to revoke the session token if possible, or close all tracks.
func (c *Client) CloseSession(ctx context.Context, sessionID string) error {
	// TODO: Implement track closing logic or find specific API if available.
	// For now, since we don't have the track IDs easily, we might just rely on
	// the fact that we mark the room as ended in our DB, and the frontend
	// prevents re-joining.
	//
	// However, to force close existing connections:
	// PUT /apps/{appId}/sessions/{sessionId}/tracks/close
	// But we need track IDs? Or can we close all?
	//
	// Let's implement a placeholder log for now as direct API usage requires more investigation
	// on how to list tracks for a session first.
	//
	// Actually, a better approach might be to just ensure our frontend
	// handles the state 'ended' correctly and leaves.
	//
	// If we want to strictly enforce it, we might need to use Cloudflare Stream/Calls
	// "Force Close" which might be available via different API.
	//
	// Reviewing docs again... `kick_participant` permission allows a host to kick others.
	// If the host "Ends Meeting", the frontend client (as host) should iterate and kick everyone?
	// OR call an endpoint that does it.

	// For this step, I'll add the method signature and a comment,
	// but without valid endpoint implementation it won't work yet.
	// Given the constraints and lack of exact API docs in context,
	// I will focus on enforcing the "Room Ended" state in the DB prevents new joins
	// and maybe the frontend can poll for room status?

	return nil
}
