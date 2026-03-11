package handler

import (
	"belaz-calendar-server/internal/models"
	"belaz-calendar-server/internal/pkg/logger"
	"belaz-calendar-server/internal/repository"
	"belaz-calendar-server/internal/service"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MaintenanceHandler struct {
	vehicleRepo repository.VehicleRepository
	typeRepo    repository.MaintenanceTypeRepository
	calendarSvc service.CalendarService
	maintSvc    service.MaintenanceService
	logger      *zap.Logger
}

func NewMaintenanceHandler(
	vehicleRepo repository.VehicleRepository,
	typeRepo repository.MaintenanceTypeRepository,
	calendarSvc service.CalendarService,
	maintSvc service.MaintenanceService,
) *MaintenanceHandler {
	return &MaintenanceHandler{
		vehicleRepo: vehicleRepo,
		typeRepo:    typeRepo,
		calendarSvc: calendarSvc,
		maintSvc:    maintSvc,
		logger:      logger.Log,
	}
}

func (h *MaintenanceHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/vehicles", h.GetVehicles)
	r.GET("/vehicles/:id", h.GetVehicle)

	r.GET("/calendar", h.GetCalendar)

	r.GET("/maintenance/types", h.GetMaintenanceTypes)

	r.GET("/maintenance/types/:id/actions", h.GetMergedActions)

	r.GET("/maintenance/:id/event-with-actions", h.GetEventWithActions)
	r.GET("/maintenance/:id/details", h.GetMaintenanceDetails)

	r.POST("/maintenance/:id/complete", h.CompleteMaintenance)
	r.PUT("/maintenance/:id/reschedule", h.RescheduleMaintenance)
	r.POST("/maintenance/seasonal", h.CreateSeasonal)
}

func (h *MaintenanceHandler) GetVehicles(c *gin.Context) {
	vehicles, err := h.vehicleRepo.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Get vehicles failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load vehicles"})
		return
	}

	h.logger.Debug("Get vehicles", zap.Int("count", len(vehicles)))
	c.JSON(http.StatusOK, gin.H{"vehicles": vehicles})
}

func (h *MaintenanceHandler) GetVehicle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("Invalid vehicle ID", zap.String("id", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	vehicle, err := h.vehicleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrVehicleNotFound) {
			h.logger.Warn("Vehicle not found", zap.Int64("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
			return
		}
		h.logger.Error("Get vehicle failed", zap.Error(err), zap.Int64("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load vehicle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"vehicle": vehicle})
}

func (h *MaintenanceHandler) GetCalendar(c *gin.Context) {
	var vehicleID *int64
	if v := c.Query("vehicle_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			vehicleID = &id
		}
	}

	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))

	h.logger.Debug("Get calendar",
		zap.Int64("vehicle_id", func() int64 {
			if vehicleID != nil {
				return *vehicleID
			}
			return 0
		}()),
		zap.Int("month", month),
		zap.Int("year", year))

	events, err := h.calendarSvc.GetEvents(c.Request.Context(), vehicleID, month, year)
	if err != nil {
		h.logger.Error("Get calendar failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load calendar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"month":  month,
		"year":   year,
	})
}

func (h *MaintenanceHandler) GetMaintenanceTypes(c *gin.Context) {
	types, err := h.typeRepo.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Get maintenance types failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load maintenance types"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"types": types})
}

func (h *MaintenanceHandler) GetMergedActions(c *gin.Context) {
	typeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type ID"})
		return
	}

	actions, err := h.maintSvc.GetMergedActions(c.Request.Context(), typeID)
	if err != nil {
		h.logger.Error("Get merged actions failed", zap.Error(err), zap.Int64("type_id", typeID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

func (h *MaintenanceHandler) CompleteMaintenance(c *gin.Context) {
	recordID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var req models.CompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid completion request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Complete maintenance requested",
		zap.Int64("record_id", recordID),
		zap.Float64("mileage", func() float64 {
			if req.Mileage != nil {
				return *req.Mileage
			}
			return 0
		}()),
		zap.Float64("hours", func() float64 {
			if req.EngineHours != nil {
				return *req.EngineHours
			}
			return 0
		}()))

	if err := h.maintSvc.Complete(c.Request.Context(), recordID, &req); err != nil {
		h.logger.Warn("Complete maintenance failed", zap.Error(err), zap.Int64("record_id", recordID))

		errorResponse := gin.H{"error": err.Error()}
		statusCode := http.StatusBadRequest

		switch {
		case err.Error() == "already completed":
			statusCode = http.StatusConflict
		case err.Error() == "mileage cannot be less than current":
			statusCode = http.StatusBadRequest
		case err.Error() == "engine hours cannot be less than current":
			statusCode = http.StatusBadRequest
		case err.Error() == "mileage must be between previous and next completed TO":
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Maintenance completed successfully",
		"record_id": recordID,
	})
}

func (h *MaintenanceHandler) RescheduleMaintenance(c *gin.Context) {
	recordID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var req models.RescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Reschedule maintenance requested",
		zap.Int64("record_id", recordID),
		zap.Time("new_date", req.NewScheduledDate))

	if err := h.maintSvc.Reschedule(c.Request.Context(), recordID, &req); err != nil {
		h.logger.Warn("Reschedule failed", zap.Error(err), zap.Int64("record_id", recordID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Maintenance rescheduled successfully",
		"record_id":          recordID,
		"new_scheduled_date": req.NewScheduledDate,
	})
}

func (h *MaintenanceHandler) CreateSeasonal(c *gin.Context) {
	var req models.SeasonalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Create seasonal TO requested",
		zap.Int64("vehicle_id", req.VehicleID),
		zap.Time("scheduled_date", req.ScheduledDate))

	record, err := h.maintSvc.CreateSeasonal(c.Request.Context(), &req)
	if err != nil {
		h.logger.Warn("Create seasonal failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Seasonal TO created/updated successfully",
		"record":  record,
	})
}

func (h *MaintenanceHandler) GetEventWithActions(c *gin.Context) {
	recordID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("Invalid record ID", zap.String("record_id", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID format"})
		return
	}

	actions, err := h.maintSvc.GetEventWithActions(c.Request.Context(), recordID)
	if err != nil {
		if errors.Is(err, repository.ErrServiceRecordNotFound) {
			h.logger.Warn("Record not found", zap.Int64("record_id", recordID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			return
		}
		h.logger.Error("Get event with actions failed", zap.Error(err), zap.Int64("record_id", recordID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load event"})
		return
	}

	h.logger.Debug("Get event with actions",
		zap.Int64("record_id", recordID),
		zap.Int("actions_count", len(actions)),
	)
	c.JSON(http.StatusOK, gin.H{
		"event": actions,
	})
}

func (h *MaintenanceHandler) GetMaintenanceDetails(c *gin.Context) {
	recordID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("Invalid record ID", zap.String("id", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	details, err := h.maintSvc.GetMaintenanceDetails(c.Request.Context(), recordID)
	if err != nil {
		if errors.Is(err, repository.ErrServiceRecordNotFound) ||
			errors.Is(err, repository.ErrVehicleNotFound) {
			h.logger.Warn("Record not found", zap.Int64("record_id", recordID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Service record not found"})
			return
		}
		h.logger.Error("Get maintenance details failed", zap.Error(err), zap.Int64("record_id", recordID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance details"})
		return
	}

	c.JSON(http.StatusOK, details)
}
