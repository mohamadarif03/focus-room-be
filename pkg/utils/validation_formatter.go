package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type SingleErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationError(err error) []SingleErrorResponse {
	var errors []SingleErrorResponse

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors = append(errors, SingleErrorResponse{Field: "general", Message: err.Error()})
		return errors
	}

	for _, fieldErr := range validationErrors {
		field := strings.ToLower(fieldErr.Field())

		errors = append(errors, SingleErrorResponse{
			Field:   field,
			Message: formatMessage(field, fieldErr.Tag(), fieldErr.Param()),
		})
	}

	return errors
}

func formatMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("The %s field is required.", field)
	case "email":
		return fmt.Sprintf("The %s field must be a valid email address.", field)
	case "min":
		return fmt.Sprintf("The %s field must have at least %s characters.", field, param)
	case "oneof":
		return fmt.Sprintf("The %s field must be one of: %s.", field, param)
	default:
		return fmt.Sprintf("The %s field is invalid (rule: %s).", field, tag)
	}
}
