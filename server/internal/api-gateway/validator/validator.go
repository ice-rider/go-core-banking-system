package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) Validate(s interface{}) []ValidationError {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []ValidationError{{Field: "", Message: err.Error()}}
	}

	var errs []ValidationError
	for _, e := range validationErrors {
		errs = append(errs, ValidationError{
			Field:   e.Field(),
			Message: formatMessage(e),
		})
	}
	return errs
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", camelToSnake(e.Field()))
	case "min":
		if e.Param() == "1" {
			return fmt.Sprintf("%s must not be empty", camelToSnake(e.Field()))
		}
		return fmt.Sprintf("%s must be at least %s characters", camelToSnake(e.Field()), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", camelToSnake(e.Field()), e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", camelToSnake(e.Field()))
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", camelToSnake(e.Field()), e.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", camelToSnake(e.Field()), e.Tag())
	}
}

func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
