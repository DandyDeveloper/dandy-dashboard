package calendar

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Event is the simplified event shape returned to the frontend.
type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Start       string `json:"start"`       // RFC3339
	End         string `json:"end"`         // RFC3339
	AllDay      bool   `json:"all_day"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// Service wraps the Google Calendar API.
type Service struct {
	cal        *gcal.Service
	calendarID string
}

func NewService(credentialsJSON, calendarID string) (*Service, error) {
	if credentialsJSON == "" {
		// Return a stub service that returns empty events when unconfigured.
		return &Service{calendarID: calendarID}, nil
	}

	var credsData []byte
	var err error

	// Support either a file path or raw JSON.
	if _, statErr := os.Stat(credentialsJSON); statErr == nil {
		credsData, err = os.ReadFile(credentialsJSON)
	} else {
		credsData = []byte(credentialsJSON)
	}
	if err != nil {
		return nil, fmt.Errorf("reading credentials: %w", err)
	}

	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, credsData, gcal.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	svc, err := gcal.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("creating calendar service: %w", err)
	}

	return &Service{cal: svc, calendarID: calendarID}, nil
}

// GetUpcomingEvents returns events starting from now up to `days` days ahead.
func (s *Service) GetUpcomingEvents(days int) ([]Event, error) {
	if s.cal == nil {
		// Return demo events when the calendar isn't configured.
		return demoEvents(), nil
	}

	if days <= 0 || days > 30 {
		days = 7
	}

	now := time.Now()
	end := now.AddDate(0, 0, days)

	list, err := s.cal.Events.List(s.calendarID).
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(50).
		Do()
	if err != nil {
		return nil, fmt.Errorf("fetching events: %w", err)
	}

	events := make([]Event, 0, len(list.Items))
	for _, item := range list.Items {
		e := Event{
			ID:          item.Id,
			Title:       item.Summary,
			Location:    item.Location,
			Description: item.Description,
			Color:       item.ColorId,
		}
		if item.Start.DateTime != "" {
			e.Start = item.Start.DateTime
			e.End = item.End.DateTime
		} else {
			e.Start = item.Start.Date
			e.End = item.End.Date
			e.AllDay = true
		}
		events = append(events, e)
	}

	return events, nil
}

func demoEvents() []Event {
	now := time.Now()
	return []Event{
		{
			ID:    "demo-1",
			Title: "Calendar not configured",
			Start: now.Add(1 * time.Hour).Format(time.RFC3339),
			End:   now.Add(2 * time.Hour).Format(time.RFC3339),
			Description: "Set GOOGLE_CREDENTIALS_JSON and GOOGLE_CALENDAR_ID in .env to show real events.",
		},
	}
}

