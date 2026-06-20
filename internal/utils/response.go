package utils

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response is the standard error/empty API response shape.
// For success responses with data, controllers use typed dto.*Response structs
// so Orval generates concrete types rather than unknown.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	// Data is only populated on generic success helpers; typed endpoints use dto envelopes.
	Data  interface{} `json:"data,omitempty"    swaggerignore:"true"`
	Error string      `json:"error,omitempty"`
	// Errors holds validation field errors; shape varies so excluded from spec.
	Errors interface{} `json:"errors,omitempty"  swaggerignore:"true"`
	Code   int         `json:"code,omitempty"`
}

// SuccessResponse sends a 200 OK response with data.
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse sends a 201 Created response with data.
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error response with the given status code.
func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, Response{
		Success: false,
		Message: message,
		Code:    status,
	})
}

// UnauthorizedResponse sends a 401 Unauthorized error.
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message)
}

// ForbiddenResponse sends a 403 Forbidden error.
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message)
}

// NotFoundResponse sends a 404 Not Found error.
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

// ConflictResponse sends a 409 Conflict error.
func ConflictResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusConflict, message)
}

// InternalServerErrorResponse sends a 500 Internal Server Error.
func InternalServerErrorResponse(c *gin.Context, err error, message string) {
	if err != nil {
		Log.WithError(err).Error(message)
	} else {
		Log.Error(message)
	}
	ErrorResponse(c, http.StatusInternalServerError, "internal server error")
}

// FormatValidationErrors converts validator.ValidationErrors to a map of field → tag.
func FormatValidationErrors(err error) map[string]string {
	errorsMap := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			errorsMap[fe.Field()] = fe.Tag()
		}
	}
	return errorsMap
}

// ValidationErrorResponse sends a 400 Bad Request with validation details.
func ValidationErrorResponse(c *gin.Context, errs interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: "validation failed",
		Errors:  errs,
		Code:    http.StatusBadRequest,
	})
}

// BindAndValidate binds JSON body and validates the struct or slice.
func BindAndValidate(c *gin.Context, req interface{}, validate *validator.Validate) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		return handleBindingError(c, err)
	}
	return handleValidationError(c, req, validate)
}

// BindAndValidateQuery binds query params and validates the struct.
func BindAndValidateQuery(c *gin.Context, req interface{}, validate *validator.Validate) bool {
	if err := c.ShouldBindQuery(req); err != nil {
		return handleBindingError(c, err)
	}
	return handleValidationError(c, req, validate)
}

func handleBindingError(c *gin.Context, err error) bool {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		ValidationErrorResponse(c, FormatValidationErrors(ve))
		return false
	}
	ValidationErrorResponse(c, err.Error())
	return false
}

func handleValidationError(c *gin.Context, req interface{}, validate *validator.Validate) bool {
	if err := validateStruct(req, validate); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			ValidationErrorResponse(c, FormatValidationErrors(ve))
			return false
		}
		ValidationErrorResponse(c, err.Error())
		return false
	}
	return true
}

func validateStruct(req interface{}, validate *validator.Validate) error {
	if validate == nil {
		return nil
	}
	value := reflect.ValueOf(req)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		return validate.Var(value.Interface(), "dive")
	default:
		return validate.Struct(value.Interface())
	}
}

// HandleAppError converts AppError into a consistent HTTP response.
func HandleAppError(c *gin.Context, err error, message string) {
	if err == nil {
		InternalServerErrorResponse(c, nil, message)
		return
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case http.StatusBadRequest:
			ErrorResponse(c, http.StatusBadRequest, appErr.Message)
		case http.StatusUnauthorized:
			UnauthorizedResponse(c, appErr.Message)
		case http.StatusForbidden:
			ForbiddenResponse(c, appErr.Message)
		case http.StatusNotFound:
			NotFoundResponse(c, appErr.Message)
		case http.StatusConflict:
			ConflictResponse(c, appErr.Message)
		case http.StatusTooManyRequests:
			ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
		default:
			InternalServerErrorResponse(c, appErr.Err, message)
		}
		return
	}
	InternalServerErrorResponse(c, err, message)
}

// HandleServiceError maps service-layer errors to HTTP responses (alias for HandleAppError).
func HandleServiceError(c *gin.Context, err error, logMessage string) {
	HandleAppError(c, err, logMessage)
}
