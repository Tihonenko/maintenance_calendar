package repository

import (
	"belaz-calendar-server/internal/models"
	"context"

	"github.com/jmoiron/sqlx"
)

type ServiceRecordItemRepository interface {
	CreateBulk(ctx context.Context, recordID int64, items []models.ChecklistItem) error
	GetByRecordID(ctx context.Context, recordID int64) ([]*models.ServiceRecordItem, error)
	GetActionsByRecordID(ctx context.Context, recordID int64) ([]*models.ServiceRecordItem, error)
	GetActionsByVehicleAndType(ctx context.Context, vehicleID, typeID int64) ([]*models.ServiceRecordItem, error)
	GetChecklistWithResults(ctx context.Context, recordID int64) ([]*models.ServiceRecordItem, error)
}

type serviceRecordItemRepository struct {
	db *sqlx.DB
}

func NewServiceRecordItemRepository(db *sqlx.DB) ServiceRecordItemRepository {
	return &serviceRecordItemRepository{db: db}
}

func (r *serviceRecordItemRepository) CreateBulk(
	ctx context.Context,
	recordID int64,
	items []models.ChecklistItem,
) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO service_record_items 
			(service_record_id, action_id, is_passed, comment, created_at)
		VALUES (:service_record_id, :action_id, :is_passed, :comment, CURRENT_TIMESTAMP)`

	for _, item := range items {
		dbItem := &struct {
			ServiceRecordID int64  `db:"service_record_id"`
			ActionID        int64  `db:"action_id"`
			IsPassed        bool   `db:"is_passed"`
			Comment         string `db:"comment"`
		}{
			ServiceRecordID: recordID,
			ActionID:        item.ActionID,
			IsPassed:        item.IsPassed,
			Comment:         item.Comment,
		}

		_, err := r.db.NamedExecContext(ctx, query, dbItem)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *serviceRecordItemRepository) GetByRecordID(
	ctx context.Context,
	recordID int64,
) ([]*models.ServiceRecordItem, error) {
	query := `
		SELECT 
			sri.id,
			sri.service_record_id,
			sri.action_id,
			sri.is_passed,
			sri.comment,
			sri.created_at,
			ma.system_node,
			ma.description
		FROM service_record_items sri
		JOIN maintenance_actions ma ON sri.action_id = ma.id
		WHERE sri.service_record_id = $1
		ORDER BY ma.sort_order`

	var items []*models.ServiceRecordItem
	err := r.db.SelectContext(ctx, &items, query, recordID)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []*models.ServiceRecordItem{}
	}

	return items, nil
}

func (r *serviceRecordItemRepository) GetActionsByRecordID(
	ctx context.Context,
	recordID int64,
) ([]*models.ServiceRecordItem, error) {
	query := `
		SELECT 
			sri.id,
			sri.service_record_id,
			sri.action_id,
			sri.is_passed,
			sri.comment,
			sri.created_at,
			ma.system_node,
			ma.description
		FROM service_record_items sri
		JOIN maintenance_actions ma ON sri.action_id = ma.id
		WHERE sri.service_record_id = $1
		ORDER BY ma.sort_order`

	var items []*models.ServiceRecordItem
	err := r.db.SelectContext(ctx, &items, query, recordID)
	if err == nil && items != nil && len(items) > 0 {
		return items, nil
	}

	getTypeQuery := `SELECT type_id FROM service_records WHERE id = $1`
	var typeID int64
	err = r.db.GetContext(ctx, &typeID, getTypeQuery, recordID)
	if err != nil {
		return nil, err
	}

	actionQuery := `
		SELECT 
			ma.id as id,
			0 as service_record_id,
			ma.id as action_id,
			FALSE as is_passed,
			'' as comment,
			CURRENT_TIMESTAMP as created_at,
			ma.system_node,
			ma.description
		FROM maintenance_actions ma
		WHERE ma.type_id = $1
		ORDER BY ma.sort_order`

	var actions []*models.ServiceRecordItem
	err = r.db.SelectContext(ctx, &actions, actionQuery, typeID)
	if err != nil {
		return nil, err
	}

	if actions == nil {
		actions = []*models.ServiceRecordItem{}
	}

	return actions, nil
}

func (r *serviceRecordItemRepository) GetActionsByVehicleAndType(
	ctx context.Context,
	vehicleID, typeID int64,
) ([]*models.ServiceRecordItem, error) {
	getLastDoneQuery := `
		SELECT id
		FROM service_records
		WHERE vehicle_id = $1 
		  AND type_id = $2
		  AND status = 'DONE'
		ORDER BY completion_date DESC
		LIMIT 1`

	var lastDoneRecordID int64
	err := r.db.GetContext(ctx, &lastDoneRecordID, getLastDoneQuery, vehicleID, typeID)

	if err == nil && lastDoneRecordID > 0 {
		query := `
			SELECT 
				sri.id,
				sri.service_record_id,
				sri.action_id,
				sri.is_passed,
				sri.comment,
				sri.created_at,
				ma.system_node,
				ma.description
			FROM service_record_items sri
			JOIN maintenance_actions ma ON sri.action_id = ma.id
			WHERE sri.service_record_id = $1
			ORDER BY ma.sort_order`

		var items []*models.ServiceRecordItem
		errQuery := r.db.SelectContext(ctx, &items, query, lastDoneRecordID)
		if errQuery != nil {
			return nil, errQuery
		}

		if items == nil {
			items = []*models.ServiceRecordItem{}
		}

		return items, nil
	}

	query := `
		SELECT 
			ma.id as id,
			0 as service_record_id,
			ma.id as action_id,
			FALSE as is_passed,
			'' as comment,
			CURRENT_TIMESTAMP as created_at,
			ma.system_node,
			ma.description
		FROM maintenance_actions ma
		WHERE ma.type_id = $1
		ORDER BY ma.sort_order`

	var items []*models.ServiceRecordItem
	errQuery := r.db.SelectContext(ctx, &items, query, typeID)
	if errQuery != nil {
		return nil, errQuery
	}

	if items == nil {
		items = []*models.ServiceRecordItem{}
	}

	return items, nil
}

func UnwrapTxFromContext(ctx context.Context) *sqlx.Tx {
	return nil
}

func (r *serviceRecordItemRepository) GetChecklistWithResults(
	ctx context.Context,
	recordID int64,
) ([]*models.ServiceRecordItem, error) {

	query := `
		SELECT 
			sri.id,
			sri.service_record_id,
			sri.action_id,
			sri.is_passed,
			sri.comment,
			sri.created_at,
			ma.system_node,
			ma.description,
			ma.sort_order
		FROM service_record_items sri
		INNER JOIN maintenance_actions ma ON sri.action_id = ma.id
		WHERE sri.service_record_id = $1
		ORDER BY ma.sort_order, sri.id
	`

	var items []*models.ServiceRecordItem
	err := r.db.SelectContext(ctx, &items, query, recordID)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []*models.ServiceRecordItem{}
	}

	return items, nil
}
