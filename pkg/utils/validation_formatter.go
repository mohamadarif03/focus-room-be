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
		return fmt.Sprintf("Kolom %s tidak boleh kosong.", field)
	case "email":
		return fmt.Sprintf("Kolom %s harus berupa alamat email yang valid.", field)
	case "min":
		return fmt.Sprintf("Kolom %s harus memiliki minimal %s karakter.", field, param)
	case "oneof":
		return fmt.Sprintf("Kolom %s harus salah satu dari: %s.", field, param)
	default:
		return fmt.Sprintf("Kolom %s tidak valid (aturan: %s).", field, tag)
	}
}