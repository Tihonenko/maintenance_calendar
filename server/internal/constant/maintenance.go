package constant

const (
	StatusPlanned   = "PLANNED"
	StatusCompleted = "COMPLETED"
	StatusCancelled = "CANCELLED"
)

const (
	UIStatusCompleted = "completed"
	UIStatusOverdue   = "overdue"
	UIStatusPending   = "pending"
	UIStatusPlanned   = "planned"
)

const (
	CyclicHourStep  = 250.0  
	CyclicHourCycle = 1000.0 

	TO1MileageStep = 5000.0
	TO2MileageStep = 10000.0
	TO3MileageStep = 20000.0
)

const (
	MileageTolerance     = 500.0
	EngineHoursTolerance = 25.0
	DefaultForecastDays  = 30
)

const (
	Hours = 5.0
	Km    = 100.0
)

const (
	TypeTO1      = "TO1"
	TypeTO2      = "TO2"
	TypeTO3      = "TO3"
	TypeSeasonal = "SEASONAL"
)
