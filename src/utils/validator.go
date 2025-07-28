package utils

import (
	"fmt"

	"github.com/Testzyler/banking-api/server/exception"
	"github.com/go-playground/validator/v10"
)

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct and returns a custom error response
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		// Convert validation errors to our custom error format
		var validationErrors []string

		for _, err := range err.(validator.ValidationErrors) {
			var message string

			switch err.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", err.Field())
			case "min":
				message = fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
			case "max":
				message = fmt.Sprintf("%s must not exceed %s", err.Field(), err.Param())
			case "email":
				message = fmt.Sprintf("%s must be a valid email address", err.Field())
			case "uuid":
				message = fmt.Sprintf("%s must be a valid UUID", err.Field())
			case "len":
				message = fmt.Sprintf("%s must be exactly %s characters", err.Field(), err.Param())
			default:
				message = fmt.Sprintf("%s is invalid", err.Field())
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

// getFieldName converts field name to a more user-friendly format
// func getFieldName(field string) string {
// 	// Convert camelCase to readable format
// 	var result strings.Builder
// 	for i, r := range field {
// 		if i > 0 && r >= 'A' && r <= 'Z' {
// 			result.WriteString(" ")
// 		}
// 		result.WriteRune(r)
// 	}

// 	return strings.ToLower(result.String())
// }

// Custom validation functions can be added here
func RegisterCustomValidations() {
	// Custom validation for account numbers (12 digits)
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

	// Custom validation for phone numbers (TH format)
	validate.RegisterValidation("phone_th", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Basic TH phone validation (starts with +, 10-12 digits)
		if len(phone) < 10 || len(phone) > 12 {
			return false
		}
		if phone[0] != '+' {
			return false
		}
		for _, char := range phone[1:] {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	})

	// Custom validation for strong passwords
	validate.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}

		hasUpper := false
		hasLower := false
		hasDigit := false
		hasSpecial := false

		for _, char := range password {
			switch {
			case char >= 'A' && char <= 'Z':
				hasUpper = true
			case char >= 'a' && char <= 'z':
				hasLower = true
			case char >= '0' && char <= '9':
				hasDigit = true
			case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
				hasSpecial = true
			}
		}

		return hasUpper && hasLower && hasDigit && hasSpecial
	})
}
