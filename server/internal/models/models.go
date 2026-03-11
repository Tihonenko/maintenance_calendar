package models

import (
	"time"
)

type PredictedTO struct {
	VehicleID      int       `json:"vehicle_id"`
	VehicleVIN     string    `json:"vehicle_vin"`
	TypeID         int       `json:"type_id"`
	TypeCode       string    `json:"type_code"`
	TypeName       string    `json:"type_name"`
	CalculatedDate time.Time `json:"calculated_date"`
	MileageTarget  float64   `json:"mileage_target"`
	HoursTarget    float64   `json:"hours_target"`
	Trigger        string    `json:"trigger"`
}

type CompletionRequest struct {
	CompletionDate string  `json:"completion_date" binding:"required"`
	Mileage        float64 `json:"mileage" binding:"required,min=0"`
	Hours          float64 `json:"hours" binding:"required,min=0"`
}
