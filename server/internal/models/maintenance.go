package models

import "time"

type MaintenanceType struct {
	ID            int64     `db:"id" json:"id"`
	Code          string    `db:"code" json:"code"`
	Name          string    `db:"name" json:"name"`
	IntervalKM    *int      `db:"interval_km" json:"interval_km"`
	IntervalHours *int      `db:"interval_hours" json:"interval_hours"`
	ParentID      *int64    `db:"parent_id" json:"parent_id"`
	IsCascading   bool      `db:"is_cascading" json:"is_cascading"`
	IsOneTime     bool      `db:"is_one_time" json:"is_one_time"`
	IsSeasonal    bool      `db:"is_seasonal" json:"is_seasonal"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type CalendarEvent struct {
	ID             int64                `db:"id" json:"id"`
	VehicleID      int64                `db:"vehicle_id" json:"vehicle_id"`
	VIN            string               `db:"vin" json:"vin"`
	TypeID         int64                `db:"type_id" json:"type_id"`
	TypeCode       string               `db:"type_code" json:"type_code"`
	TypeName       string               `db:"type_name" json:"type_name"`
	Status         string               `db:"status" json:"status"`
	CalculatedDate time.Time            `db:"calculated_date" json:"calculated_date"`
	ScheduledDate  time.Time            `db:"scheduled_date" json:"scheduled_date"`
	UIStatus       string               `db:"-" json:"ui_status"`
	DaysOverdue    int                  `db:"-" json:"days_overdue"`
	IsRescheduled  bool                 `db:"is_rescheduled" json:"is_rescheduled"`
	Actions        []*ServiceRecordItem `db:"-" json:"actions,omitempty"`
}

type ServiceRecord struct {
	ID                  int64      `db:"id" json:"id"`
	VehicleID           int64      `db:"vehicle_id" json:"vehicle_id"`
	TypeID              int64      `db:"type_id" json:"type_id"`
	Status              string     `db:"status" json:"status"`
	CalculatedDate      time.Time  `db:"calculated_date" json:"calculated_date"`
	ScheduledDate       time.Time  `db:"scheduled_date" json:"scheduled_date"`
	CompletionDate      *time.Time `db:"completion_date" json:"completion_date,omitempty"`
	MileageAtCompletion *float64   `db:"mileage_at_completion" json:"mileage_at_completion,omitempty"`
	HoursAtCompletion   *float64   `db:"hours_at_completion" json:"hours_at_completion,omitempty"`
	IsRescheduled       bool       `db:"is_rescheduled" json:"is_rescheduled"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`

	VIN             string `db:"vin" json:"vin"`
	MaintenanceType string `db:"maintenance_type" json:"maintenance_type"` // Code from MaintenanceType
	DaysOverdue     int    `db:"-" json:"days_overdue"`                    // Для сортировки
}

type ServiceRecordItem struct {
	ID              int64     `db:"id" json:"id"`
	ServiceRecordID int64     `db:"service_record_id" json:"service_record_id"`
	ActionID        int64     `db:"action_id" json:"action_id"`
	IsPassed        bool      `db:"is_passed" json:"is_passed"`
	Comment         string    `db:"comment" json:"comment,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"
	`
	SystemNode  string `db:"system_node" json:"system_node"`
	Description string `db:"description" json:"description"`
	SortOrder   int    `db:"sort_order" json:"sort_order"`
}

type MaintenanceAction struct {
	ID          int64     `db:"id" json:"id"`
	TypeID      int64     `db:"type_id" json:"type_id"`
	SystemNode  string    `db:"system_node" json:"system_node"`
	Description string    `db:"description" json:"description"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type CompleteRequest struct {
	CompletionDate time.Time       `json:"completion_date" binding:"required"`
	Mileage        *float64        `json:"mileage" binding:"omitempty"`
	EngineHours    *float64        `json:"engine_hours" binding:"omitempty"`
	Checklist      []ChecklistItem `json:"checklist,omitempty"`
}

type ChecklistItem struct {
	ActionID int64  `json:"action_id" binding:"required"`
	IsPassed bool   `json:"is_passed"`
	Comment  string `json:"comment,omitempty"`
}

type RescheduleRequest struct {
	NewScheduledDate time.Time `json:"new_scheduled_date" binding:"required"`
}

type SeasonalRequest struct {
	VehicleID     int64     `json:"vehicle_id" binding:"required"`
	ScheduledDate time.Time `json:"scheduled_date" binding:"required"`
}

type MaintenanceDetailsResponse struct {
	RecordID       int64      `json:"record_id"`
	VehicleID      int64      `json:"vehicle_id"`
	VIN            string     `json:"vin"`
	TypeCode       string     `json:"type_code"`
	TypeName       string     `json:"type_name"`
	Status         string     `json:"status"`
	CalculatedDate time.Time  `json:"calculated_date"`
	ScheduledDate  time.Time  `json:"scheduled_date"`
	CompletionDate *time.Time `json:"completion_date,omitempty"`
	Mileage        *float64   `json:"mileage_at_completion,omitempty"`
	EngineHours    *float64   `json:"engine_hours_at_completion,omitempty"`
	IsCascading    bool       `json:"is_cascading"`
	IsRescheduled  bool       `json:"is_rescheduled"`

	Items []MaintenanceItemDetail `json:"items"`
}

type MaintenanceItemDetail struct {
	ActionID    int64  `json:"action_id"`
	SystemNode  string `json:"system_node"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`

	IsPassed    *bool   `json:"is_passed,omitempty"`
	Comment     *string `json:"comment,omitempty"`
	IsCompleted bool    `json:"is_completed"`
}
