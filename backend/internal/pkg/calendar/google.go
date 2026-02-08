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
	ID          string
}

type Service interface {
	CreateEvent(ctx context.Context, token *oauth2.Token, event Event) (string, string, error) // Returns ID, Link, Error
	UpdateEvent(ctx context.Context, token *oauth2.Token, eventID string, event Event) error
	DeleteEvent(ctx context.Context, token *oauth2.Token, eventID string) error
	GetAuthURL(state string) string
	ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error)
	GetBusyTimes(ctx context.Context, token *oauth2.Token, start, end time.Time) ([]TimePeriod, error)
	ListEvents(ctx context.Context, token *oauth2.Token, start, end time.Time) ([]Event, error)
}

type TimePeriod struct {
	Start time.Time
	End   time.Time
}

type googleService struct {
	config *oauth2.Config
}

func NewGoogleService(clientID, clientSecret, redirectURL string) Service {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}
	return &googleService{config: config}
}

func (s *googleService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *googleService) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

func (s *googleService) CreateEvent(ctx context.Context, token *oauth2.Token, evt Event) (string, string, error) {
	client := s.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", "", err
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
				RequestId: "meet-clone-" + time.Now().String(),
			},
		},
	}

	createdEvent, err := srv.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		return "", "", err
	}
	return createdEvent.Id, createdEvent.HtmlLink, nil
}

func (s *googleService) UpdateEvent(ctx context.Context, token *oauth2.Token, eventID string, evt Event) error {
	client := s.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
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
	}

	_, err = srv.Events.Update("primary", eventID, event).Context(ctx).Do()
	return err
}

func (s *googleService) DeleteEvent(ctx context.Context, token *oauth2.Token, eventID string) error {
	client := s.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	return srv.Events.Delete("primary", eventID).Context(ctx).Do()
}

func (s *googleService) GetBusyTimes(ctx context.Context, token *oauth2.Token, start, end time.Time) ([]TimePeriod, error) {
	client := s.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	freebusyReq := &calendar.FreeBusyRequest{
		TimeMin: start.Format(time.RFC3339),
		TimeMax: end.Format(time.RFC3339),
		Items:   []*calendar.FreeBusyRequestItem{{Id: "primary"}},
	}

	fbResponse, err := srv.Freebusy.Query(freebusyReq).Do()
	if err != nil {
		return nil, err
	}

	var busyTimes []TimePeriod
	if calendarBusy, ok := fbResponse.Calendars["primary"]; ok {
		for _, busy := range calendarBusy.Busy {
			start, _ := time.Parse(time.RFC3339, busy.Start)
			end, _ := time.Parse(time.RFC3339, busy.End)
			busyTimes = append(busyTimes, TimePeriod{Start: start, End: end})
		}
	}

	return busyTimes, nil
}

func (s *googleService) ListEvents(ctx context.Context, token *oauth2.Token, start, end time.Time) ([]Event, error) {
	client := s.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	events, err := srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	var result []Event
	for _, item := range events.Items {
		var start, end time.Time
		if item.Start.DateTime != "" {
			start, _ = time.Parse(time.RFC3339, item.Start.DateTime)
		} else {
			start, _ = time.Parse("2006-01-02", item.Start.Date)
		}
		if item.End.DateTime != "" {
			end, _ = time.Parse(time.RFC3339, item.End.DateTime)
		} else {
			end, _ = time.Parse("2006-01-02", item.End.Date)
		}

		result = append(result, Event{
			Summary:     item.Summary,
			Description: item.Description,
			Start:       start,
			End:         end,
			ID:          item.Id,
		})
	}
	return result, nil
}
