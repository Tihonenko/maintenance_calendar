package service

import (
	"belaz-calendar-server/internal/models"
	"math"
	"time"
)

type MaintenanceCalculator interface {
	CalculateNextDue(vehicle *models.Vehicle, mType *models.MaintenanceType, lastCompleted *models.ServiceRecord) *CalcResult
	CalculateNextCyclicDue(vehicle *models.Vehicle, nextType *models.MaintenanceType, lastCompleted *models.ServiceRecord) *CyclicCalcResult
	IsWithinTolerance(actual, expected, tolerance float64) bool
}

type CalcResult struct {
	NextMileage  float64
	NextHours    float64
	DueDate      time.Time
	TriggeredBy  string
	DaysUntilDue int
}

type CyclicCalcResult struct {
	DueDate      time.Time
	DaysUntilDue int
}

type maintenanceCalculator struct{}

func NewMaintenanceCalculator() MaintenanceCalculator {
	return &maintenanceCalculator{}
}

func (c *maintenanceCalculator) CalculateNextCyclicDue(
	vehicle *models.Vehicle,
	nextType *models.MaintenanceType,
	lastCompleted *models.ServiceRecord,
) *CyclicCalcResult {
	const defaultForecastDays = 30

	var intervalKM, intervalHours float64

	switch nextType.Code {
	case "TO1":
		intervalKM = 5000.0
		intervalHours = 250.0
	case "TO2":
		intervalKM = 10000.0
		intervalHours = 500.0
	case "TO3":
		intervalKM = 20000.0
		intervalHours = 1000.0
	default:
		intervalHours = 250.0
		if nextType.IntervalKM != nil && *nextType.IntervalKM > 0 {
			intervalKM = float64(*nextType.IntervalKM)
		}
	}

	currentMileage := vehicle.TotalMileage
	currentHours := vehicle.TotalEngineHours

	var dateByMileage, dateByHours *time.Time

	if intervalKM > 0 {
		nextMileageThreshold := math.Ceil(currentMileage/intervalKM) * intervalKM
		if nextMileageThreshold <= currentMileage {
			nextMileageThreshold += intervalKM
		}

		remainingKM := nextMileageThreshold - currentMileage
		if remainingKM > 0 && vehicle.AvgSpeed > 0 {
			hoursToGo := remainingKM / vehicle.AvgSpeed
			due := time.Now().Add(time.Duration(hoursToGo) * time.Hour)
			dateByMileage = &due
		}
	}

	if intervalHours > 0 {
		nextHoursThreshold := math.Ceil(currentHours/intervalHours) * intervalHours
		if nextHoursThreshold <= currentHours {
			nextHoursThreshold += intervalHours
		}

		remainingHours := nextHoursThreshold - currentHours
		if remainingHours > 0 {
			due := time.Now().Add(time.Duration(remainingHours) * time.Hour)
			dateByHours = &due
		}
	}

	var dueDate time.Time
	switch {
	case dateByMileage != nil && dateByHours != nil:
		if dateByMileage.Before(*dateByHours) {
			dueDate = *dateByMileage
		} else {
			dueDate = *dateByHours
		}
	case dateByMileage != nil:
		dueDate = *dateByMileage
	case dateByHours != nil:
		dueDate = *dateByHours
	default:
		dueDate = time.Now().Add(defaultForecastDays * 24 * time.Hour)
	}

	daysUntil := int(math.Ceil(dueDate.Sub(time.Now()).Hours() / 24))
	if daysUntil < 0 {
		daysUntil = 0
	}

	return &CyclicCalcResult{
		DueDate:      dueDate,
		DaysUntilDue: daysUntil,
	}
}

func (c *maintenanceCalculator) IsWithinTolerance(actual, expected, tolerance float64) bool {
	diff := math.Abs(actual - expected)
	return diff <= tolerance
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
		if nextThreshold <= currentMileage {
			nextThreshold += interval
		}
		thresholdMileage = &nextThreshold
		triggeredBy = "mileage"
	}

	if mType.IntervalHours != nil && *mType.IntervalHours > 0 {
		interval := float64(*mType.IntervalHours)
		nextThreshold := math.Ceil(currentHours/interval) * interval
		if nextThreshold <= currentHours {
			nextThreshold += interval
		}
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

