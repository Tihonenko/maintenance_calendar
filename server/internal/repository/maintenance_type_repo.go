package repository

import (
	"belaz-calendar-server/internal/models"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrMaintenanceTypeNotFound = errors.New("maintenance type not found")

type MaintenanceTypeRepository interface {
	GetAll(ctx context.Context) ([]*models.MaintenanceType, error)
	GetByID(ctx context.Context, id int64) (*models.MaintenanceType, error)
	GetByCode(ctx context.Context, code string) (*models.MaintenanceType, error)
	GetLastCompletedCyclicType(ctx context.Context, vehicleID int64) (string, error)
}

type maintenanceTypeRepository struct {
	db *sqlx.DB
}

func NewMaintenanceTypeRepository(db *sqlx.DB) MaintenanceTypeRepository {
	return &maintenanceTypeRepository{db: db}
}

func (r *maintenanceTypeRepository) GetAll(ctx context.Context) ([]*models.MaintenanceType, error) {

	query := `SELECT * FROM maintenance_types ORDER BY id`

	var types []*models.MaintenanceType

	err := r.db.SelectContext(ctx, &types, query)

	if err != nil {
		return nil, err
	}

	return types, nil

}

func (r *maintenanceTypeRepository) GetByID(ctx context.Context, id int64) (*models.MaintenanceType, error) {
	query := `SELECT * FROM maintenance_types WHERE id = $1`

	mType := &models.MaintenanceType{}

	err := r.db.GetContext(ctx, mType, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMaintenanceTypeNotFound
		}
		return nil, err
	}

	return mType, nil
}

func (r *maintenanceTypeRepository) GetByCode(ctx context.Context, code string) (*models.MaintenanceType, error) {
	query := `SELECT * FROM maintenance_types WHERE code = $1`

	mType := &models.MaintenanceType{}

	err := r.db.GetContext(ctx, mType, query, code)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMaintenanceTypeNotFound
		}
		return nil, err
	}

	return mType, nil
}

func (r *maintenanceTypeRepository) GetLastCompletedCyclicType(
	ctx context.Context,
	vehicleID int64,
) (string, error) {

	query := `
		SELECT mt.code
		FROM service_records sr
		JOIN maintenance_types mt ON sr.type_id = mt.id
		WHERE sr.vehicle_id = $1
		  AND mt.code IN ('TO1', 'TO2', 'TO3')
		  AND sr.status = 'DONE'
		  AND sr.completion_date IS NOT NULL
		ORDER BY sr.completion_date DESC
		LIMIT 1
	`

	var code string
	err := r.db.GetContext(ctx, &code, query, vehicleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return code, nil
}
