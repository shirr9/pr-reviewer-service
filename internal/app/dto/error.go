package dto

// ErrorResponse represents the error response structure.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new ErrorResponse.
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}
