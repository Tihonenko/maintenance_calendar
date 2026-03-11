package service

import (
	"belaz-calendar-server/internal/models"
	"math"
	"time"
)

type MaintenanceCalculator interface {
	CalculateNextDue(vehicle *models.Vehicle, mType *models.MaintenanceType, lastCompleted *models.ServiceRecord) *CalcResult
}

type CalcResult struct {
	NextMileage  float64
	NextHours    float64
	DueDate      time.Time
	TriggeredBy  string
	DaysUntilDue int
}

type maintenanceCalculator struct{}

func NewMaintenanceCalculator() MaintenanceCalculator {
	return &maintenanceCalculator{}
}
func (c *maintenanceCalculator) CalculateNextDue(
	vehicle *models.Vehicle,
	mType *models.MaintenanceType,
	lastCompleted *models.ServiceRecord,
) *CalcResult {

	currentMileage := vehicle.TotalMileage
	currentHours := vehicle.TotalEngineHours

	if (mType.IntervalKM == nil || *mType.IntervalKM <= 0) &&
		(mType.IntervalHours == nil || *mType.IntervalHours <= 0) {
		return nil
	}

	var thresholdMileage, thresholdHours *float64
	var triggeredBy string

	if mType.IntervalKM != nil && *mType.IntervalKM > 0 {
		interval := float64(*mType.IntervalKM)
		nextThreshold := math.Ceil(currentMileage/interval) * interval
		thresholdMileage = &nextThreshold
		triggeredBy = "mileage"
	}

	if mType.IntervalHours != nil && *mType.IntervalHours > 0 {
		interval := float64(*mType.IntervalHours)
		nextThreshold := math.Ceil(currentHours/interval) * interval
		thresholdHours = &nextThreshold
		if triggeredBy == "" {
			triggeredBy = "hours"
		}
	}

	var dateByMileage, dateByHours *time.Time

	if thresholdMileage != nil {
		remaining := *thresholdMileage - currentMileage
		if remaining <= 0 {
			dateByMileage = &time.Time{}
		} else {
			if vehicle.AvgSpeed > 0 {
				hoursToGo := remaining / vehicle.AvgSpeed
				due := time.Now().Add(time.Duration(hoursToGo) * time.Hour)
				dateByMileage = &due
			}
		}
	}

	if thresholdHours != nil {
		remaining := *thresholdHours - currentHours
		if remaining <= 0 {
			dateByHours = &time.Time{}
		} else {
			due := time.Now().Add(time.Duration(remaining) * time.Hour)
			dateByHours = &due
		}
	}

	var dueDate time.Time
	switch {
	case dateByMileage != nil && dateByHours != nil:

		if dateByMileage.IsZero() || (!dateByHours.IsZero() && dateByMileage.Before(*dateByHours)) {
			dueDate = *dateByMileage
			triggeredBy = "mileage"
		} else {
			dueDate = *dateByHours
			triggeredBy = "hours"
		}
	case dateByMileage != nil:
		dueDate = *dateByMileage
		triggeredBy = "mileage"
	case dateByHours != nil:
		dueDate = *dateByHours
		triggeredBy = "hours"
	default:
		return nil
	}

	daysUntil := int(math.Ceil(dueDate.Sub(time.Now()).Hours() / 24))
	if daysUntil < 0 {
		daysUntil = 0
	}

	return &CalcResult{
		NextMileage: func() float64 {
			if thresholdMileage != nil {
				return *thresholdMileage
			}
			return 0
		}(),
		NextHours: func() float64 {
			if thresholdHours != nil {
				return *thresholdHours
			}
			return 0
		}(),
		DueDate:      dueDate,
		TriggeredBy:  triggeredBy,
		DaysUntilDue: daysUntil,
	}
}
