package models

import "time"

type Vehicle struct {
	ID               int64     `db:"id" json:"id"`
	VIN              string    `db:"vin" json:"vin"`
	TotalMileage     float64   `db:"total_mileage" json:"total_mileage"`
	TotalEngineHours float64   `db:"total_engine_hours" json:"total_engine_hours"`
	AvgSpeed         float64   `db:"avg_speed" json:"avg_speed"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type CreateVehicleRequest struct {
	VIN              string  `json:"vin" binding:"required,len=17"`
	TotalMileage     float64 `json:"total_mileage" binding:"min=0"`
	TotalEngineHours float64 `json:"total_engine_hours" binding:"min=0"`
	AvgSpeed         float64 `json:"avg_speed" binding:"min=10.5,max=16.3"`
}

type UpdateVehicleRequest struct {
	TotalMileage     *float64 `json:"total_mileage,omitempty" binding:"omitempty,min=0"`
	TotalEngineHours *float64 `json:"total_engine_hours,omitempty" binding:"omitempty,min=0"`
	AvgSpeed         *float64 `json:"avg_speed,omitempty" binding:"omitempty,min=10.5,max=16.3"`
}
