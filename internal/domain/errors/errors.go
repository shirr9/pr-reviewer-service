package errors

const (
	CodeTeamExists  = "TEAM_EXISTS"
	CodePRExists    = "PR_EXISTS"
	CodePRMerged    = "PR_MERGED"
	CodeNotAssigned = "NOT_ASSIGNED"
	CodeNoCandidate = "NO_CANDIDATE"
	CodeNotFound    = "NOT_FOUND"
)

// AppError represents a domain error with code and message.
type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// New creates a new AppError.
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewTeamExists(message string) *AppError {
	return New(CodeTeamExists, message)
}

func NewPRExists(message string) *AppError {
	return New(CodePRExists, message)
}

func NewPRMerged(message string) *AppError {
	return New(CodePRMerged, message)
}

func NewNotAssigned(message string) *AppError {
	return New(CodeNotAssigned, message)
}

func NewNoCandidate(message string) *AppError {
	return New(CodeNoCandidate, message)
}

func NewNotFound(message string) *AppError {
	return New(CodeNotFound, message)
}
