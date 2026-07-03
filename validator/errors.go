package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors converts validator.ValidationErrors to a map of field -> message
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, ve := range validationErrors {
			field := toSnakeCase(ve.Field())
			errors[field] = getValidationMessage(ve)
		}
	}
	return errors
}

func getValidationMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + ve.Param() + " characters"
	case "max":
		return "must be at most " + ve.Param() + " characters"
	case "oneof":
		return "must be one of: " + ve.Param()
	case "uuid":
		return "must be a valid UUID"
	default:
		return "invalid value for field " + ve.Field()
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
