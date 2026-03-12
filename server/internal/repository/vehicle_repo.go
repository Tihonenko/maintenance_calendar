package repository

import (
	"belaz-calendar-server/internal/models"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrVehicleNotFound = errors.New("vehicle not found")
	ErrVehicleExists   = errors.New("vehicle with this VIN alredy exists")
)

type VehicleRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Vehicle, error)
	GetAll(ctx context.Context) ([]*models.Vehicle, error)
	UpdateCurrentMetrics(ctx context.Context, id int64, mileage, hours float64) error
}

type vehicleRepository struct {
	db *sqlx.DB
}

func NewVehicleRepository(db *sqlx.DB) VehicleRepository {
	return &vehicleRepository{db: db}
}

func (r *vehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	query := `
		INSERT INTO vehicles (vin, total_mileage, total_engine_hours, avg_speed)
			VALUES (:vin, :total_mileage, :total_engine_hours, :avg_speed)
				RETURNING id`
	row, err := r.db.NamedQueryContext(ctx, query, vehicle)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVehicleNotFound
		}
		return err
	}

	defer row.Close()

	if row.Next() {
		if err := row.StructScan(vehicle); err != nil {
			return err
		}
	}

	return row.Err()
}

func (r *vehicleRepository) GetByID(ctx context.Context, id int64) (*models.Vehicle, error) {
	query := `SELECT * FROM vehicles WHERE id = $1`

	vehicle := &models.Vehicle{}

	err := r.db.GetContext(ctx, vehicle, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}

		return nil, err
	}

	return vehicle, nil
}

func (r *vehicleRepository) GetAll(ctx context.Context) ([]*models.Vehicle, error) {
	query := `SELECT * FROM vehicles ORDER BY vin`

	var vehicle []*models.Vehicle

	err := r.db.SelectContext(ctx, &vehicle, query)

	if err != nil {
		return nil, err
	}

	return vehicle, nil
}

func (r *vehicleRepository) UpdateCurrentMetrics(ctx context.Context, id int64, mileage, hours float64) error {

	query := `UPDATE vehicles
				SET total_mileage = $1, total_engine_hours = $2, updated_at = CURRENT_TIMESTAMP
					WHERE id = $3`

	res, err := r.db.ExecContext(ctx, query, mileage, hours, id)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrVehicleNotFound
	}

	return nil
}
