package errors

type AppError struct {
	Err        error
	Code       string
	StatusCode int
	Field      string
}

func (e *AppError) Error() string {
	return e.Err.Error()
}

// Helper function to check if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// Helper function to get the status code from an error
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return 500
}

// Helper function to get the error code from an error
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "ERROR"
}
