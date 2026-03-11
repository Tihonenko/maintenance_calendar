package repository

import (
	"belaz-calendar-server/internal/models"
	"context"

	"github.com/jmoiron/sqlx"
)

type maintenanceActionRepository struct {
	db *sqlx.DB
}

func NewMaintenanceActionRepository(db *sqlx.DB) MaintenanceActionRepository {
	return &maintenanceActionRepository{db: db}
}

func (r *maintenanceActionRepository) GetByTypeIDs(ctx context.Context, typesIDs []int64) ([]*models.MaintenanceAction, error) {
	if len(typesIDs) == 0 {
		return []*models.MaintenanceAction{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT * FROM maintenance_actions 
			WHERE type_id IN (?) 
				ORDER BY sort_order`,
		typesIDs)

	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var actions []*models.MaintenanceAction

	err = r.db.SelectContext(ctx, &actions, query, args...)

	if err != nil {
		return nil, err
	}

	if actions == nil {
		actions = []*models.MaintenanceAction{}
	}

	return actions, nil
}
