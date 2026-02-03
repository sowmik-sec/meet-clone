package calendar

import (
	"context"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Event struct {
	Summary     string
	Description string
	Start       time.Time
	End         time.Time
	Attendees   []string
}

type Service interface {
	CreateEvent(ctx context.Context, token *oauth2.Token, event Event) (string, error)
	GetAuthURL() string
	ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error)
}

type googleService struct {
	config *oauth2.Config
}

func NewGoogleService(clientID, clientSecret, redirectURL string) Service {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{calendar.CalendarEventsScope},
		Endpoint:     google.Endpoint,
	}
	return &googleService{config: config}
}

func (s *googleService) GetAuthURL() string {
	return s.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (s *googleService) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

func (s *googleService) CreateEvent(ctx context.Context, token *oauth2.Token, evt Event) (string, error) {
	client := s.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	attendees := make([]*calendar.EventAttendee, len(evt.Attendees))
	for i, email := range evt.Attendees {
		attendees[i] = &calendar.EventAttendee{Email: email}
	}

	event := &calendar.Event{
		Summary:     evt.Summary,
		Description: evt.Description,
		Start: &calendar.EventDateTime{
			DateTime: evt.Start.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: evt.End.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Attendees: attendees,
		ConferenceData: &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: "meet-clone-" + time.Now().String(), // Unique ID
				// ConferenceSolutionKey: &calendar.ConferenceSolutionKey{Type: "hangoutsMeet"}, // Optional: Integrate Google Meet? No, we use our link.
			},
		},
	}

	// We'll put our meeting link in the description or location
	// For "hangoutsMeet" we need to be a real partner or something.
	// Simplest: just add link to description/location.

	createdEvent, err := srv.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return createdEvent.HtmlLink, nil
}
