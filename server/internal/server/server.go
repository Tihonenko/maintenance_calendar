package server

import (
	"belaz-calendar-server/internal/handler"
	"belaz-calendar-server/internal/pkg/logger"
	"belaz-calendar-server/internal/repository"
	"belaz-calendar-server/internal/service"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Server struct {
	router       *gin.Engine
	httpSrv      *http.Server
	db           *sqlx.DB
	maintService service.MaintenanceService
}

func NewServer(db *sqlx.DB) *Server {
	logger.Init()

	vehicleRepo := repository.NewVehicleRepository(db)
	maintenanceTypeRepo := repository.NewMaintenanceTypeRepository(db)
	maintenanceActionRepo := repository.NewMaintenanceActionRepository(db)
	serviceRecordRepo := repository.NewServiceRecordRepository(db)
	serviceRecordItemRepo := repository.NewServiceRecordItemRepository(db)

	calculator := service.NewMaintenanceCalculator()
	calendarService := service.NewCalendarService(serviceRecordRepo, calculator)

	maintenanceService := service.NewMaintenanceService(
		maintenanceTypeRepo,
		maintenanceActionRepo,
		serviceRecordRepo,
		serviceRecordItemRepo,
		vehicleRepo,
		calculator,
	)

	maintenanceHandler := handler.NewMaintenanceHandler(
		vehicleRepo,
		maintenanceTypeRepo,
		calendarService,
		maintenanceService,
	)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(loggingMiddleware())
	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		maintenanceHandler.RegisterRoutes(api)
	}

	httpSrv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Log.Info("Server initialized successfully",
		zap.String("version", "1.0.0"),
		zap.Int("repositories", 5),
		zap.Int("services", 2),
		zap.Int("handlers", 1))

	return &Server{
		router:       router,
		httpSrv:      httpSrv,
		db:           db,
		maintService: maintenanceService,
	}
}

func (s *Server) Run(port string) error {
	s.httpSrv.Addr = port
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) Close() error {
	return s.db.Close()
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Log.Info("HTTP Request",
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *Server) InitializeMaintenanceSchedules(ctx context.Context) error {
	if s.maintService == nil {
		return errors.New("maintenance service not initialized")
	}
	return s.maintService.InitializeMaintenanceSchedules(ctx)
}
