package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/common"
	"github.com/nguereza-tony/corekit/errors"
	"github.com/nguereza-tony/corekit/logger"
)

// ResponseDTO represents a standard API response
type ResponseDTO[T any] struct {
	Success   bool   `json:"success" example:"true"`
	Timestamp string `json:"timestamp" example:"2026-06-25T10:15:30Z"`
	Code      string `json:"code" example:"OK"`
	RequestID string `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Data      T      `json:"data,omitempty"`
	Meta      any    `json:"meta,omitempty"`
}

// ErrorResponseDTO represents a standard API error response
type ErrorResponseDTO struct {
	Success   bool              `json:"success" example:"false"`
	Timestamp string            `json:"timestamp" example:"2026-06-25T10:15:30Z"`
	Code      string            `json:"code" example:"VALIDATION_ERROR"`
	Message   string            `json:"message,omitempty" example:"Request processing error"`
	RequestID string            `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Errors    map[string]string `json:"errors,omitempty"`
	Meta      any               `json:"meta,omitempty"`
}

// NewResponseDTO creates a new ResponseDTO with timestamp and default code
func NewResponseDTO[T any](success bool) ResponseDTO[T] {
	code := "OK"
	if !success {
		code = "ERROR"
	}
	return ResponseDTO[T]{
		Success:   success,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Code:      code,
	}
}

// WithRequestID adds a request ID to the response
func (r ResponseDTO[T]) WithRequestID(requestID string) ResponseDTO[T] {
	r.RequestID = requestID
	return r
}

// WithData adds data to the response
func (r ResponseDTO[T]) WithData(data T) ResponseDTO[T] {
	r.Data = data
	return r
}

// WithMeta adds metadata to the response
func (r ResponseDTO[T]) WithMeta(meta any) ResponseDTO[T] {
	r.Meta = meta
	return r
}

// WithCode sets a custom code for the response
func (r ResponseDTO[T]) WithCode(code string) ResponseDTO[T] {
	r.Code = code
	return r
}

// ErrorResponseDTO builder methods
func NewErrorResponseDTO() ErrorResponseDTO {
	return ErrorResponseDTO{
		Success:   false,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (r ErrorResponseDTO) WithRequestID(requestID string) ErrorResponseDTO {
	r.RequestID = requestID
	return r
}

func (r ErrorResponseDTO) WithMessage(message string) ErrorResponseDTO {
	r.Message = message
	return r
}

func (r ErrorResponseDTO) WithCode(code string) ErrorResponseDTO {
	r.Code = code
	return r
}

func (r ErrorResponseDTO) WithErrors(errors map[string]string) ErrorResponseDTO {
	r.Errors = errors
	return r
}

func (r ErrorResponseDTO) WithMeta(meta any) ErrorResponseDTO {
	r.Meta = meta
	return r
}

// ============ SUCCESS RESPONSES ============

func OK[T any](
	c *gin.Context,
	statusCode int,
	data *T,
	code string,
	meta any,
) {
	apiResponse(
		c,
		true,
		statusCode,
		data,
		code,
		map[string]string{},
		"",
		meta,
	)
}

func Success[T any](
	c *gin.Context,
	data *T,
	meta any,
) {
	OK(c, http.StatusOK, data, "OK", meta)
}

func Created[T any](
	c *gin.Context,
	data *T,
	meta any,
) {
	OK(c, http.StatusCreated, data, "CREATED", meta)
}

func Accepted[T any](
	c *gin.Context,
	data *T,
	meta any,
) {
	OK(c, http.StatusAccepted, data, "ACCEPTED", meta)
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// ============ ERROR RESPONSES ============

func Error(
	c *gin.Context,
	statusCode int,
	errors map[string]string,
	code string,
	message string,
	meta any,
) {
	apiResponse[any](
		c,
		false,
		statusCode,
		nil,
		code,
		errors,
		message,
		meta,
	)
}

// 400 Bad Request
func BadRequest(c *gin.Context, code string, message string) {
	if code == "" {
		code = "BAD_REQUEST"
	}
	Error(
		c,
		http.StatusBadRequest,
		map[string]string{},
		code,
		message,
		nil,
	)
}

// 401 Unauthorized
func Unauthorized(c *gin.Context, code string, message string) {
	if code == "" {
		code = "UNAUTHORIZED_ACCESS"
	}
	Error(
		c,
		http.StatusUnauthorized,
		map[string]string{},
		code,
		message,
		nil,
	)
}

// 403 Forbidden
func Forbidden(c *gin.Context, code string, message string) {
	if code == "" {
		code = "FORBIDDEN"
	}
	Error(
		c,
		http.StatusForbidden,
		map[string]string{},
		code,
		message,
		nil,
	)
}

// 404 Not Found
func NotFound(c *gin.Context, code string, message string) {
	if code == "" {
		code = "RESOURCE_NOT_FOUND"
	}
	Error(
		c,
		http.StatusNotFound,
		map[string]string{},
		code,
		message,
		nil,
	)
}

// 409 Conflict
func Conflict(c *gin.Context, code string, errors map[string]string) {
	if code == "" {
		code = "DUPLICATE_RESOURCE"
	}
	Error(
		c,
		http.StatusConflict,
		errors,
		code,
		"",
		nil,
	)
}

// 422 Unprocessable Entity
func InputValidationError(c *gin.Context, errors map[string]string, message string) {
	if message == "" {
		message = "Invalid Request Parameter(s)"
	}
	Error(
		c,
		http.StatusUnprocessableEntity,
		errors,
		"INVALID_INPUT",
		message,
		nil,
	)
}

// 429 Too Many Requests
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "Too many requests. Please try again later."
	}
	Error(
		c,
		http.StatusTooManyRequests,
		map[string]string{},
		"TOO_MANY_REQUESTS",
		message,
		nil,
	)
}

// 500 Internal Server Error
func InternalServerError(c *gin.Context, code, message string) {
	if code == "" {
		code = "INTERNAL_SERVER_ERROR"
	}
	Error(
		c,
		http.StatusInternalServerError,
		map[string]string{},
		code,
		message,
		nil,
	)
}

// ============ PRIVATE FUNCTIONS ============

func apiResponse[T any](
	c *gin.Context,
	success bool,
	statusCode int,
	data *T,
	code string,
	errors map[string]string,
	message string,
	meta any,
) {
	requestID := common.GetRequestID(c)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	if success {
		res := ResponseDTO[T]{
			Success:   success,
			Timestamp: timestamp,
			Code:      code,
			RequestID: requestID,
			Meta:      meta,
		}
		if data != nil {
			res.Data = *data
		}
		c.JSON(statusCode, res)
	} else {
		res := ErrorResponseDTO{
			Success:   success,
			Timestamp: timestamp,
			Code:      code,
			RequestID: requestID,
			Meta:      meta,
			Message:   message,
		}
		if len(errors) > 0 {
			res.Errors = errors
		}
		c.JSON(statusCode, res)
	}
}

func HandleError(c *gin.Context, logger *logger.Logger, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		if appErr.StatusCode >= 500 {
			logger.Errorf("An unexpected error occurred: %v", err)
		}

		fields := map[string]string{}
		if appErr.Field != "" {
			fields[appErr.Field] = appErr.Error()
		}

		Error(
			c,
			appErr.StatusCode,
			fields,
			appErr.Code,
			appErr.Error(),
			nil,
		)
		return
	}

	// Fallback for unknow error (not AppError)
	logger.Errorf("An unexpected error occurred: %v", err)
	InternalServerError(c, "", "An unexpected error occurred")
}
