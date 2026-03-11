package service

import (
	"belaz-calendar-server/internal/models"
	"belaz-calendar-server/internal/repository"
	"context"
	"time"
)

type CalendarService interface {
	GetEvents(ctx context.Context, vehicleID *int64, month, year int) ([]*models.CalendarEvent, error)
	calculateUIStatus(event *models.CalendarEvent, today time.Time) (string, int)
}

type calendarService struct {
	recordRepo repository.ServiceRecordRepository
	calculator MaintenanceCalculator
}

func NewCalendarService(
	recordRepo repository.ServiceRecordRepository,
	calculator MaintenanceCalculator,
) CalendarService {
	return &calendarService{recordRepo: recordRepo, calculator: calculator}
}

func (s *calendarService) GetEvents(
	ctx context.Context,
	vehicleID *int64,
	month, year int,
) ([]*models.CalendarEvent, error) {

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	events, err := s.recordRepo.GetCalendarEvents(ctx, vehicleID, start, end)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)

	for _, e := range events {
		e.UIStatus, e.DaysOverdue = s.calculateUIStatus(e, today)
	}

	return events, nil
}

func (s *calendarService) calculateUIStatus(e *models.CalendarEvent, today time.Time) (string, int) {
	if e.Status == "DONE" {
		return "light-green", 0
	}

	scheduled := e.ScheduledDate.Truncate(24 * time.Hour)
	today = today.Truncate(24 * time.Hour)

	if scheduled.After(today) {
		return "green", 0
	}

	daysOverdue := int(today.Sub(scheduled).Hours() / 24)

	if daysOverdue > 5 {
		return "red", daysOverdue
	}

	return "orange", daysOverdue
}
