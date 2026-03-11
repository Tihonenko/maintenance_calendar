package repository

import (
	"belaz-calendar-server/internal/models"
	"context"
)

type VehicleRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Vehicle, error)
	GetAll(ctx context.Context) ([]*models.Vehicle, error)
	UpdateCurrentMetrics(ctx context.Context, id int64, mileage, hours float64) error
}

type MaintenanceTypeRepository interface {
	GetAll(ctx context.Context) ([]*models.MaintenanceType, error)
	GetByID(ctx context.Context, id int64) (*models.MaintenanceType, error)
	GetByCode(ctx context.Context, code string) (*models.MaintenanceType, error)
	GetLastCompletedCyclicType(ctx context.Context, vehicleID int64) (string, error)
}

type MaintenanceActionRepository interface {
	GetByTypeIDs(ctx context.Context, typeIDs []int64) ([]*models.MaintenanceAction, error)
}
