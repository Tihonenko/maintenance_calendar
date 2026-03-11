package repository

import (
	"belaz-calendar-server/internal/models"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrServiceRecordNotFound = errors.New("service record not found")

type ServiceRecordRepository interface {
	GetCalendarEvents(ctx context.Context, vehicleID *int64, start, end time.Time) ([]*models.CalendarEvent, error)

	GetByID(ctx context.Context, id int64) (*models.ServiceRecord, error)
	Create(ctx context.Context, record *models.ServiceRecord) error
	Update(ctx context.Context, record *models.ServiceRecord) error

	GetLastCompleted(ctx context.Context, vehicleID, typeID int64) (*models.ServiceRecord, error)
	GetNextPlanned(ctx context.Context, vehicleID, typeID int64) (*models.ServiceRecord, error)
	GetActiveSeasonal(ctx context.Context, vehicleID int64) (*models.ServiceRecord, error)
	HasActiveRecord(ctx context.Context, vehicleID, typeID int64) (bool, error)

	BeginTx(ctx context.Context) (Tx, error)
}

type serviceRecordRepository struct {
	db *sqlx.DB
}

func NewServiceRecordRepository(db *sqlx.DB) ServiceRecordRepository {
	return &serviceRecordRepository{db: db}
}

func (r *serviceRecordRepository) GetCalendarEvents(
	ctx context.Context,
	vehicleID *int64,
	start, end time.Time,
) ([]*models.CalendarEvent, error) {

	query := `
		SELECT 
			sr.id,
			sr.vehicle_id,
			v.vin,
			sr.type_id,
			mt.code as type_code,
			mt.name as type_name,
			sr.status,
			sr.calculated_date,
			sr.scheduled_date,
			sr.is_rescheduled
		FROM service_records sr
		JOIN vehicles v ON sr.vehicle_id = v.id
		JOIN maintenance_types mt ON sr.type_id = mt.id
		WHERE sr.scheduled_date >= $1 
		  AND sr.scheduled_date < $2`

	args := []interface{}{start, end}

	if vehicleID != nil {
		query += " AND sr.vehicle_id = $3"
		args = append(args, *vehicleID)
	}

	query += " ORDER BY sr.scheduled_date, sr.vehicle_id"

	var events []*models.CalendarEvent
	err := r.db.SelectContext(ctx, &events, query, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *serviceRecordRepository) GetByID(ctx context.Context, id int64) (*models.ServiceRecord, error) {
	query := `SELECT id, vehicle_id, type_id, status, calculated_date, scheduled_date, 
					 completion_date, mileage_at_completion, hours_at_completion, 
					 is_rescheduled, created_at, updated_at 
			  FROM service_records WHERE id = $1`

	record := &models.ServiceRecord{}
	err := r.db.GetContext(ctx, record, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrServiceRecordNotFound
		}
		return nil, err
	}

	return record, nil
}

func (r *serviceRecordRepository) Create(ctx context.Context, record *models.ServiceRecord) error {
	query := `
		INSERT INTO service_records 
			(vehicle_id, type_id, status, calculated_date, scheduled_date, is_rescheduled)
		VALUES (:vehicle_id, :type_id, :status, :calculated_date, :scheduled_date, :is_rescheduled)
		RETURNING id, created_at, updated_at`

	row, err := r.db.NamedQueryContext(ctx, query, record)
	if err != nil {
		return err
	}
	defer row.Close()

	if row.Next() {
		err = row.StructScan(record)
		if err != nil {
			return err
		}
	}

	return row.Err()
}

func (r *serviceRecordRepository) Update(ctx context.Context, record *models.ServiceRecord) error {
	query := `
		UPDATE service_records 
		SET status = :status,
		    calculated_date = :calculated_date,
		    scheduled_date = :scheduled_date,
		    completion_date = :completion_date,
		    mileage_at_completion = :mileage_at_completion,
		    hours_at_completion = :hours_at_completion,
		    is_rescheduled = :is_rescheduled,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id`

	res, err := r.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrServiceRecordNotFound
	}

	return nil
}

func (r *serviceRecordRepository) GetLastCompleted(
	ctx context.Context,
	vehicleID, typeID int64,
) (*models.ServiceRecord, error) {
	query := `
		SELECT id, vehicle_id, type_id, status, calculated_date, scheduled_date, 
		       completion_date, mileage_at_completion, hours_at_completion, 
		       is_rescheduled, created_at, updated_at 
		FROM service_records 
		WHERE vehicle_id = $1 
		  AND type_id = $2 
		  AND status = 'DONE' 
		  AND completion_date IS NOT NULL
		ORDER BY completion_date DESC 
		LIMIT 1`

	record := &models.ServiceRecord{}
	err := r.db.GetContext(ctx, record, query, vehicleID, typeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func (r *serviceRecordRepository) GetNextPlanned(
	ctx context.Context,
	vehicleID, typeID int64,
) (*models.ServiceRecord, error) {
	query := `
		SELECT id, vehicle_id, type_id, status, calculated_date, scheduled_date, 
		       mileage_at_completion, hours_at_completion
		FROM service_records 
		WHERE vehicle_id = $1 
		  AND type_id = $2 
		  AND status IN ('PLANNED', 'OVERDUE')
		ORDER BY scheduled_date ASC 
		LIMIT 1`

	record := &models.ServiceRecord{}
	err := r.db.GetContext(ctx, record, query, vehicleID, typeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func (r *serviceRecordRepository) GetActiveSeasonal(
	ctx context.Context,
	vehicleID int64,
) (*models.ServiceRecord, error) {
	query := `
		SELECT sr.id, sr.vehicle_id, sr.type_id, sr.status, sr.calculated_date, 
		       sr.scheduled_date, sr.is_rescheduled, sr.created_at, sr.updated_at 
		FROM service_records sr
		JOIN maintenance_types mt ON sr.type_id = mt.id
		WHERE sr.vehicle_id = $1 
		  AND mt.is_seasonal = TRUE 
		  AND sr.status IN ('PLANNED', 'OVERDUE')
		ORDER BY sr.scheduled_date ASC 
		LIMIT 1`

	record := &models.ServiceRecord{}
	err := r.db.GetContext(ctx, record, query, vehicleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func (r *serviceRecordRepository) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx: tx}, nil
}

func (r *serviceRecordRepository) HasActiveRecord(
	ctx context.Context,
	vehicleID, typeID int64,
) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM service_records 
			WHERE vehicle_id = $1 
			  AND type_id = $2 
			  AND status IN ('PLANNED', 'OVERDUE')
			LIMIT 1
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, vehicleID, typeID)
	if err != nil {
		return false, err
	}

	return exists, nil
}
