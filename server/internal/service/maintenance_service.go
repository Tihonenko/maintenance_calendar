package service

// ТО3 = ТО1+ТО2+ТО3, ТО2 = ТО1+ТО2

import (
	"belaz-calendar-server/internal/models"
	"belaz-calendar-server/internal/pkg/logger"
	"belaz-calendar-server/internal/repository"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

var (
	ErrPastDate           = errors.New("cannot schedule in the past")
	ErrWindowExceeded     = errors.New("reschedule allowed only ±2 days from calculated date")
	ErrMileageBackward    = errors.New("mileage cannot be less than current")
	ErrHoursBackward      = errors.New("engine hours cannot be less than current")
	ErrMileageOutOfBounds = errors.New("mileage must be between previous and next completed TO")
	ErrHoursOutOfBounds   = errors.New("engine hours must be between previous and next completed TO")
	ErrCompletionInFuture = errors.New("completion date cannot be in the future")
)

type MaintenanceService interface {
	GetMaintenanceTypes(ctx context.Context) ([]*models.MaintenanceType, error)
	GetVehicles(ctx context.Context) ([]*models.Vehicle, error)
	GetVehicleByID(ctx context.Context, id int64) (*models.Vehicle, error)
	GetMergedActions(ctx context.Context, typeID int64) ([]*models.MaintenanceAction, error)
	GetEventWithActions(ctx context.Context, recordID int64) ([]*models.MaintenanceAction, error)
	Complete(ctx context.Context, recordID int64, req *models.CompleteRequest) error
	Reschedule(ctx context.Context, recordID int64, req *models.RescheduleRequest) error
	CreateSeasonal(ctx context.Context, req *models.SeasonalRequest) (*models.ServiceRecord, error)
	GetMaintenanceDetails(ctx context.Context, recordID int64) (*models.MaintenanceDetailsResponse, error)
	InitializeMaintenanceSchedules(ctx context.Context) error
}

type maintenanceService struct {
	typeRepo    repository.MaintenanceTypeRepository
	actionRepo  repository.MaintenanceActionRepository
	recordRepo  repository.ServiceRecordRepository
	itemRepo    repository.ServiceRecordItemRepository
	vehicleRepo repository.VehicleRepository
	calculator  MaintenanceCalculator
	logger      *zap.Logger
}

func NewMaintenanceService(
	typeRepo repository.MaintenanceTypeRepository,
	actionRepo repository.MaintenanceActionRepository,
	recordRepo repository.ServiceRecordRepository,
	itemRepo repository.ServiceRecordItemRepository,
	vehicleRepo repository.VehicleRepository,
	calculator MaintenanceCalculator,
) MaintenanceService {
	return &maintenanceService{
		typeRepo: typeRepo, actionRepo: actionRepo, recordRepo: recordRepo,
		itemRepo: itemRepo, vehicleRepo: vehicleRepo, calculator: calculator,
		logger: logger.Log,
	}
}

func (s *maintenanceService) GetMergedActions(
	ctx context.Context,
	typeID int64,
) ([]*models.MaintenanceAction, error) {

	mType, err := s.typeRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, err
	}

	if !mType.IsCascading {
		return s.actionRepo.GetByTypeIDs(ctx, []int64{typeID})
	}

	typeIDs := []int64{typeID}

	if mType.Code == "TO3" {
		if t2, _ := s.typeRepo.GetByCode(ctx, "TO2"); t2 != nil {
			typeIDs = append([]int64{t2.ID}, typeIDs...)
			if t1, _ := s.typeRepo.GetByCode(ctx, "TO1"); t1 != nil {
				typeIDs = append([]int64{t1.ID}, typeIDs...)
			}
		}
	} else if mType.Code == "TO2" {
		if t1, _ := s.typeRepo.GetByCode(ctx, "TO1"); t1 != nil {
			typeIDs = append([]int64{t1.ID}, typeIDs...)
		}
	}

	return s.actionRepo.GetByTypeIDs(ctx, typeIDs)
}

func (s *maintenanceService) Complete(
	ctx context.Context,
	recordID int64,
	req *models.CompleteRequest,
) error {

	tx, err := s.recordRepo.BeginTx(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	record, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("failed to get record %d: %w", recordID, err)
	}
	if record.Status == "DONE" {
		return errors.New("already completed")
	}

	vehicle, err := s.vehicleRepo.GetByID(ctx, record.VehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle %d: %w", record.VehicleID, err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	completionDateTrunc := req.CompletionDate.UTC().Truncate(24 * time.Hour)
	if completionDateTrunc.After(today) {
		return ErrCompletionInFuture
	}

	mileage := vehicle.TotalMileage
	engineHours := vehicle.TotalEngineHours

	if req.Mileage != nil {
		mileage = *req.Mileage
	}
	if req.EngineHours != nil {
		engineHours = *req.EngineHours
	}

	if mileage < vehicle.TotalMileage {
		return fmt.Errorf("mileage cannot decrease: %f < %f", mileage, vehicle.TotalMileage)
	}
	if engineHours < vehicle.TotalEngineHours {
		return fmt.Errorf("engine hours cannot decrease: %f < %f", engineHours, vehicle.TotalEngineHours)
	}

	now := time.Now()
	record.Status = "DONE"
	record.CompletionDate = &req.CompletionDate
	record.MileageAtCompletion = &mileage
	record.HoursAtCompletion = &engineHours
	record.UpdatedAt = now

	if err := s.recordRepo.Update(ctx, record); err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	if len(req.Checklist) > 0 {
		if err := s.itemRepo.CreateBulk(ctx, recordID, req.Checklist); err != nil {
			return fmt.Errorf("failed to create checklist: %w", err)
		}
	}

	if err := s.vehicleRepo.UpdateCurrentMetrics(ctx, vehicle.ID, mileage, engineHours); err != nil {
		return fmt.Errorf("failed to update vehicle metrics: %w", err)
	}

	currentType, err := s.typeRepo.GetByID(ctx, record.TypeID)
	if err != nil {
		s.logger.Error("Failed to get current maintenance type",
			zap.Error(err),
			zap.Int64("type_id", record.TypeID))
		return fmt.Errorf("failed to get maintenance type: %w", err)
	}

	if !currentType.IsSeasonal && !currentType.IsOneTime {

		nextTypeCode := determineNextCyclicType(currentType.Code)

		nextType, err := s.typeRepo.GetByCode(ctx, nextTypeCode)
		if err != nil {
			s.logger.Error("Failed to get NEXT maintenance type",
				zap.Error(err),
				zap.String("current_type", currentType.Code),
				zap.String("next_type_code", nextTypeCode))
			return fmt.Errorf("next type '%s' not found: %w", nextTypeCode, err)
		}

		s.logger.Debug("Calculating next maintenance",
			zap.String("current_type", currentType.Code),
			zap.String("next_type", nextTypeCode),
			zap.Int64("vehicle_id", vehicle.ID),
			zap.Float64("current_mileage", vehicle.TotalMileage),
			zap.Float64("current_hours", vehicle.TotalEngineHours),
			zap.Float64("avg_speed", vehicle.AvgSpeed),
			zap.Int("next_interval_km", func() int {
				if nextType.IntervalKM != nil {
					return *nextType.IntervalKM
				}
				return 0
			}()),
			zap.Int("next_interval_hours", func() int {
				if nextType.IntervalHours != nil {
					return *nextType.IntervalHours
				}
				return 0
			}()))

		result := s.calculator.CalculateNextDue(vehicle, nextType, record)

		if result == nil {
			s.logger.Warn("Calculator returned nil - no next maintenance scheduled",
				zap.Int64("vehicle_id", vehicle.ID),
				zap.String("current_type", currentType.Code),
				zap.String("next_type", nextType.Code))
		} else {
			s.logger.Info("Calculator result",
				zap.String("next_type", nextType.Code),
				zap.Time("due_date", result.DueDate),
				zap.Float64("mileage_threshold", result.NextMileage),
				zap.Float64("hours_threshold", result.NextHours),
				zap.String("triggered_by", result.TriggeredBy),
				zap.Int("days_until_due", result.DaysUntilDue))

			nextRecord := &models.ServiceRecord{
				VehicleID:      vehicle.ID,
				TypeID:         nextType.ID,
				Status:         "PLANNED",
				CalculatedDate: result.DueDate,
				ScheduledDate:  result.DueDate,
				IsRescheduled:  false,
			}

			if err := s.recordRepo.Create(ctx, nextRecord); err != nil {
				s.logger.Error("Failed to CREATE next maintenance record",
					zap.Error(err),
					zap.Int64("vehicle_id", vehicle.ID),
					zap.String("next_type", nextType.Code),
					zap.Any("record", nextRecord))
			} else {
				s.logger.Info("Next maintenance record CREATED",
					zap.Int64("record_id", nextRecord.ID),
					zap.Int64("vehicle_id", vehicle.ID),
					zap.String("type", nextType.Code),
					zap.Time("scheduled_date", result.DueDate),
					zap.Float64("threshold_mileage", result.NextMileage),
					zap.Float64("threshold_hours", result.NextHours))
			}
		}
	}

	s.logger.Info("Maintenance completed successfully",
		zap.Int64("record_id", recordID),
		zap.Int64("vehicle_id", vehicle.ID),
		zap.String("type", currentType.Code),
		zap.Float64("mileage", mileage),
		zap.Float64("hours", engineHours))

	return tx.Commit()
}
func (s *maintenanceService) Reschedule(
	ctx context.Context,
	recordID int64,
	req *models.RescheduleRequest,
) error {

	record, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}

	if req.NewScheduledDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return ErrPastDate
	}

	mType, err := s.typeRepo.GetByID(ctx, record.TypeID)
	if err != nil {
		return err
	}

	if !mType.IsSeasonal {
		calc := record.CalculatedDate.Truncate(24 * time.Hour)
		newD := req.NewScheduledDate.Truncate(24 * time.Hour)
		diff := newD.Sub(calc).Hours() / 24

		if diff < -2 || diff > 2 {
			return fmt.Errorf("%w: diff=%.1f days", ErrWindowExceeded, diff)
		}
	}

	record.ScheduledDate = req.NewScheduledDate
	record.IsRescheduled = true
	record.UpdatedAt = time.Now()

	return s.recordRepo.Update(ctx, record)
}

func (s *maintenanceService) CreateSeasonal(
	ctx context.Context,
	req *models.SeasonalRequest,
) (*models.ServiceRecord, error) {

	if req.ScheduledDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, ErrPastDate
	}

	seasonalType, err := s.typeRepo.GetByCode(ctx, "SEASONAL")
	if err != nil {
		return nil, err
	}

	active, _ := s.recordRepo.GetActiveSeasonal(ctx, req.VehicleID)

	if active != nil {
		active.ScheduledDate = req.ScheduledDate
		active.CalculatedDate = req.ScheduledDate
		active.IsRescheduled = true
		err := s.recordRepo.Update(ctx, active)
		return active, err
	}

	record := &models.ServiceRecord{
		VehicleID:      req.VehicleID,
		TypeID:         seasonalType.ID,
		Status:         "PLANNED",
		CalculatedDate: req.ScheduledDate,
		ScheduledDate:  req.ScheduledDate,
	}

	err = s.recordRepo.Create(ctx, record)
	return record, err
}

func (s *maintenanceService) GetMaintenanceTypes(ctx context.Context) ([]*models.MaintenanceType, error) {
	return s.typeRepo.GetAll(ctx)
}

func (s *maintenanceService) GetVehicles(ctx context.Context) ([]*models.Vehicle, error) {
	return s.vehicleRepo.GetAll(ctx)
}

func (s *maintenanceService) GetVehicleByID(ctx context.Context, id int64) (*models.Vehicle, error) {
	return s.vehicleRepo.GetByID(ctx, id)
}

func (s *maintenanceService) GetEventWithActions(ctx context.Context, recordID int64) ([]*models.MaintenanceAction, error) {
	record, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	mType, err := s.typeRepo.GetByID(ctx, record.TypeID)
	if err != nil {
		return nil, err
	}

	actions, err := s.GetMergedActions(ctx, mType.ID)
	if err != nil {
		actions = []*models.MaintenanceAction{}
	}

	return actions, nil
}

func (s *maintenanceService) GetMaintenanceDetails(
	ctx context.Context,
	recordID int64,
) (*models.MaintenanceDetailsResponse, error) {

	record, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	vehicle, err := s.vehicleRepo.GetByID(ctx, record.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle: %w", err)
	}

	mType, err := s.typeRepo.GetByID(ctx, record.TypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance type: %w", err)
	}

	response := &models.MaintenanceDetailsResponse{
		RecordID:       record.ID,
		VehicleID:      record.VehicleID,
		VIN:            vehicle.VIN,
		TypeCode:       mType.Code,
		TypeName:       mType.Name,
		Status:         record.Status,
		CalculatedDate: record.CalculatedDate,
		ScheduledDate:  record.ScheduledDate,
		CompletionDate: record.CompletionDate,
		Mileage:        record.MileageAtCompletion,
		EngineHours:    record.HoursAtCompletion,
		IsCascading:    mType.IsCascading,
		IsRescheduled:  record.IsRescheduled,
		Items:          []models.MaintenanceItemDetail{},
	}
	var actions []*models.MaintenanceAction
	if mType.IsCascading {
		actions, err = s.GetMergedActions(ctx, record.TypeID)
	} else {
		actions, err = s.actionRepo.GetByTypeIDs(ctx, []int64{record.TypeID})
	}
	if err != nil {
		s.logger.Warn("Failed to get actions", zap.Error(err), zap.Int64("record_id", recordID))
		actions = []*models.MaintenanceAction{}
	}

	completedMap := make(map[int64]*models.ServiceRecordItem)
	if record.Status == "DONE" {
		checklistItems, err := s.itemRepo.GetChecklistWithResults(ctx, recordID)
		if err != nil {
			s.logger.Warn("Failed to get checklist", zap.Error(err), zap.Int64("record_id", recordID))
		} else {
			for _, item := range checklistItems {
				completedMap[item.ActionID] = item
			}
		}
	}

	for _, action := range actions {
		itemDetail := models.MaintenanceItemDetail{
			ActionID:    action.ID,
			SystemNode:  action.SystemNode,
			Description: action.Description,
			SortOrder:   action.SortOrder,
			IsPassed:    nil,
			Comment:     nil,
			IsCompleted: false,
		}
		if record.Status == "DONE" {
			if completedItem, found := completedMap[action.ID]; found {

				passed := completedItem.IsPassed
				comment := completedItem.Comment

				itemDetail.IsPassed = &passed
				itemDetail.Comment = &comment
				itemDetail.IsCompleted = true
			} else {

				s.logger.Debug("Action not found in checklist",
					zap.Int64("action_id", action.ID),
					zap.Int64("record_id", recordID))
			}
		}

		response.Items = append(response.Items, itemDetail)
	}

	sort.Slice(response.Items, func(i, j int) bool {
		return response.Items[i].SortOrder < response.Items[j].SortOrder
	})

	return response, nil
}

func (s *maintenanceService) InitializeMaintenanceSchedules(ctx context.Context) error {
	s.logger.Info("Starting maintenance schedule initialization...")

	// 1. Получаем все автомобили
	vehicles, err := s.vehicleRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get vehicles for initialization", zap.Error(err))
		return fmt.Errorf("failed to get vehicles: %w", err)
	}

	s.logger.Info("Found vehicles to check", zap.Int("count", len(vehicles)))

	// 2. Получаем все типы ТО
	allTypes, err := s.typeRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get maintenance types", zap.Error(err))
		return err
	}

	createdCount := 0

	// 3. Для каждого автомобиля
	for _, vehicle := range vehicles {
		s.logger.Debug("Checking vehicle", zap.Int64("id", vehicle.ID), zap.String("vin", vehicle.VIN))

		lastCyclicCode, err := s.typeRepo.GetLastCompletedCyclicType(ctx, vehicle.ID)
		if err != nil {
			s.logger.Warn("Failed to get last cyclic TO",
				zap.Int64("vehicle_id", vehicle.ID),
				zap.Error(err))
		}

		nextCyclicCode := determineNextCyclicType(lastCyclicCode)

		nextCyclicType, err := s.typeRepo.GetByCode(ctx, nextCyclicCode)
		if err != nil {
			s.logger.Warn("Failed to get next cyclic type",
				zap.String("code", nextCyclicCode),
				zap.Error(err))
		} else {
			hasPlanned, err := s.recordRepo.HasActiveRecord(ctx, vehicle.ID, nextCyclicType.ID)
			if err != nil {
				s.logger.Warn("Failed to check existing record",
					zap.Int64("vehicle_id", vehicle.ID),
					zap.String("type_code", nextCyclicCode))
			} else if !hasPlanned {
				result := s.calculator.CalculateNextDue(vehicle, nextCyclicType, nil)
				if result != nil {
					record := &models.ServiceRecord{
						VehicleID:      vehicle.ID,
						TypeID:         nextCyclicType.ID,
						Status:         "PLANNED",
						CalculatedDate: result.DueDate,
						ScheduledDate:  result.DueDate,
						IsRescheduled:  false,
					}
					if err := s.recordRepo.Create(ctx, record); err != nil {
						s.logger.Error("Failed to create cyclic maintenance record",
							zap.Error(err),
							zap.Int64("vehicle_id", vehicle.ID),
							zap.String("type_code", nextCyclicCode))
					} else {
						createdCount++
						s.logger.Info("Cyclic maintenance record created",
							zap.Int64("record_id", record.ID),
							zap.Int64("vehicle_id", vehicle.ID),
							zap.String("type_code", nextCyclicCode),
							zap.Time("scheduled_date", result.DueDate))
					}
				}
			}
		}

		for _, mType := range allTypes {
			if isCyclicType(mType.Code) || mType.IsSeasonal || mType.IsOneTime {
				continue
			}

			hasPlanned, err := s.recordRepo.HasActiveRecord(ctx, vehicle.ID, mType.ID)
			if err != nil {
				s.logger.Warn("Failed to check existing record for DTO",
					zap.Int64("vehicle_id", vehicle.ID),
					zap.String("type_code", mType.Code))
				continue
			}

			if !hasPlanned {
				result := s.calculator.CalculateNextDue(vehicle, mType, nil)
				if result != nil {
					record := &models.ServiceRecord{
						VehicleID:      vehicle.ID,
						TypeID:         mType.ID,
						Status:         "PLANNED",
						CalculatedDate: result.DueDate,
						ScheduledDate:  result.DueDate,
						IsRescheduled:  false,
					}
					if err := s.recordRepo.Create(ctx, record); err != nil {
						s.logger.Error("Failed to create DTO maintenance record",
							zap.Error(err),
							zap.Int64("vehicle_id", vehicle.ID),
							zap.String("type_code", mType.Code))
					} else {
						createdCount++
						s.logger.Info("✓ DTO maintenance record created",
							zap.Int64("record_id", record.ID),
							zap.Int64("vehicle_id", vehicle.ID),
							zap.String("type_code", mType.Code))
					}
				}
			}
		}
	}

	s.logger.Info("Maintenance schedule initialization completed",
		zap.Int("vehicles_checked", len(vehicles)),
		zap.Int("records_created", createdCount))

	return nil
}

func determineNextCyclicType(lastTypeCode string) string {
	switch lastTypeCode {
	case "TO1":
		return "TO2"
	case "TO2":
		return "TO3"
	case "TO3":
		return "TO1"
	default:
		return "TO1"
	}
}

func isCyclicType(code string) bool {
	return code == "TO1" || code == "TO2" || code == "TO3"
}
