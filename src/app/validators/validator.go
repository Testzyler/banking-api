package validators

import (
	"fmt"
	"strings"

	"github.com/Testzyler/banking-api/server/exception"
	"github.com/go-playground/validator/v10"
)

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			var message string

			switch err.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", getFieldName(err.Field()))
			case "min":
				message = fmt.Sprintf("%s must be at least %s", getFieldName(err.Field()), err.Param())
			case "max":
				message = fmt.Sprintf("%s must not exceed %s", getFieldName(err.Field()), err.Param())
			case "email":
				message = fmt.Sprintf("%s must be a valid email address", getFieldName(err.Field()))
			case "uuid":
				message = fmt.Sprintf("%s must be a valid UUID", getFieldName(err.Field()))
			case "len":
				message = fmt.Sprintf("%s must be exactly %s characters", getFieldName(err.Field()), err.Param())
			case "numeric":
				message = fmt.Sprintf("%s must be a number", getFieldName(err.Field()))
			// Custom validation error messages
			case "account_number":
				message = fmt.Sprintf("%s must be exactly 12 digits", getFieldName(err.Field()))
			default:
				message = fmt.Sprintf("%s is invalid", getFieldName(err.Field()))
			}

			validationErrors = append(validationErrors, message)
		}

		return exception.NewValidationError(map[string]interface{}{
			"errors":  validationErrors,
			"message": "Validation failed for the provided data",
		})
	}

	return nil
}

// Convert camelCase to readable format
func getFieldName(field string) string {
	var result strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteString(" ")
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// Custom validation functions
func RegisterCustomValidations() {
	// Custom validation for account numbers (7-12 digits)
	validate.RegisterValidation("account_number", func(fl validator.FieldLevel) bool {
		accountNo := fl.Field().String()
		if len(accountNo) != 12 {
			return false
		}
		for _, char := range accountNo {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	})
}
