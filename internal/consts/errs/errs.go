package errs

type Errorf struct {
	Type      string
	Message   string
	Error     error
	ReturnRaw bool
}

// Generic Errors
const (
	ErrInternalServer  = "INTERNAL_SERVER_ERROR"
	ErrBadRequest      = "BAD_REQUEST"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrForbidden       = "FORBIDDEN"
	ErrNotFound        = "NOT_FOUND"
	ErrConflict        = "CONFLICT"
	ErrTooManyRequests = "TOO_MANY_REQUESTS"
	ErrEnvNotFound     = "ENV_NOT_FOUND"
)

// Validation & Input Errors
const (
	ErrInvalidInput  = "INVALID_INPUT"
	ErrMissingField  = "MISSING_FIELD"
	ErrBadForm       = "BAD_FORM"
	ErrInvalidFormat = "INVALID_FORMAT"
	ErrOutOfRange    = "OUT_OF_RANGE"
)

// Authentication & Authorization Errors
const (
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrTokenExpired       = "TOKEN_EXPIRED"
	ErrTokenInvalid       = "TOKEN_INVALID"
	ErrPermissionDenied   = "PERMISSION_DENIED"
)

// Database Errors
const (
	ErrDBConnection = "DATABASE_CONNECTION_ERROR"
	ErrDBQuery      = "DATABASE_QUERY_ERROR"
	ErrDBNotFound   = "DATABASE_RECORD_NOT_FOUND"
	ErrDBConflict   = "DATABASE_CONFLICT"
	ErrDBTimeout    = "DATABASE_TIMEOUT"
)

// PG errors
const (
	// Constraint Violations
	PGErrUniqueViolation          = "23505" // Duplicate key (e.g., duplicate email)
	PGErrForeignKeyViolation      = "23503" // Referential integrity error
	PGErrNotNullViolation         = "23502" // A required column is null
	PGErrCheckConstraintViolation = "23514" // A CHECK constraint failed
	PGErrExclusionViolation       = "23P01" // An EXCLUDE constraint failed

	// Data Format Errors
	PGErrInvalidTextRepresentation = "22P02" // Bad format (e.g., invalid UUID)
	PGErrDivisionByZero            = "22012" // Division by zero in a query

	// Syntax & Reference Errors
	PGErrSyntaxError       = "42601" // SQL syntax mistake
	PGErrInvalidForeignKey = "42830" // Foreign key points to wrong type

	// Deadlocks & Locking Errors
	PGErrDeadlockDetected = "40P01" // Two transactions waiting on each other

	// Not Found
	PGErrNoRowsFound = "no rows in result set"
)

// Networking & API Errors
const (
	ErrTimeout            = "REQUEST_TIMEOUT"
	ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrGatewayTimeout     = "GATEWAY_TIMEOUT"
	ErrRateLimited        = "RATE_LIMIT_EXCEEDED"
)

// File & Storage Errors
const (
	ErrFileNotFound        = "FILE_NOT_FOUND"
	ErrFileTooLarge        = "FILE_TOO_LARGE"
	ErrStorageFailed       = "STORAGE_OPERATION_FAILED"
	ErrInsufficientStorage = "INSUFFICIENT_STORAGE"
)

// Custom Business Logic Errors
const (
	ErrActionNotAllowed = "ACTION_NOT_ALLOWED"
	ErrResourceLocked   = "RESOURCE_LOCKED"
	ErrDependencyFailed = "DEPENDENCY_FAILED"
	ErrStateConflict    = "STATE_CONFLICT"
)
