package validators

import (
	"testing"

	"github.com/Testzyler/banking-api/server/response"
)

// Test struct for account number validation
type AccountTest struct {
	AccountNumber string `validate:"account_number"`
}

// Test struct for phone validation
type PhoneTest struct {
	Phone string `validate:"phone_th"`
}

// Test struct for password validation
type PasswordTest struct {
	Password string `validate:"strong_password"`
}

func TestAccountNumberValidator(t *testing.T) {
	RegisterCustomValidations()

	tests := []struct {
		name          string
		accountNumber string
		expectValid   bool
	}{
		{
			name:          "Valid 12-digit account number",
			accountNumber: "123456789012",
			expectValid:   true,
		},
		{
			name:          "Valid 12-digit account number with zeros",
			accountNumber: "000000000000",
			expectValid:   true,
		},
		{
			name:          "Invalid - too short (11 digits)",
			accountNumber: "12345678901",
			expectValid:   false,
		},
		{
			name:          "Invalid - too long (13 digits)",
			accountNumber: "1234567890123",
			expectValid:   false,
		},
		{
			name:          "Invalid - contains letters",
			accountNumber: "12345678901a",
			expectValid:   false,
		},
		{
			name:          "Invalid - contains special characters",
			accountNumber: "123456789-12",
			expectValid:   false,
		},
		{
			name:          "Invalid - empty string",
			accountNumber: "",
			expectValid:   false,
		},
		{
			name:          "Invalid - contains spaces",
			accountNumber: "123 456 789012",
			expectValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := AccountTest{AccountNumber: tt.accountNumber}
			err := ValidateStruct(account)

			if tt.expectValid && err != nil {
				t.Errorf("Expected valid account number %s, but got error: %v", tt.accountNumber, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid account number %s, but validation passed", tt.accountNumber)
			}
		})
	}
}

func TestPhoneThValidator(t *testing.T) {
	RegisterCustomValidations()

	tests := []struct {
		name        string
		phone       string
		expectValid bool
	}{
		{
			name:        "Valid Thai phone - 10 digits with +",
			phone:       "+1234567890",
			expectValid: true,
		},
		{
			name:        "Valid Thai phone - 11 digits with +",
			phone:       "+12345678901",
			expectValid: true,
		},
		{
			name:        "Valid Thai phone - 12 digits with +",
			phone:       "+123456789012",
			expectValid: true,
		},
		{
			name:        "Valid Thai phone - realistic TH number",
			phone:       "+66812345678",
			expectValid: true,
		},
		{
			name:        "Invalid - no + prefix",
			phone:       "1234567890",
			expectValid: false,
		},
		{
			name:        "Invalid - too short (9 digits with +)",
			phone:       "+123456789",
			expectValid: false,
		},
		{
			name:        "Invalid - too long (13 digits with +)",
			phone:       "+1234567890123",
			expectValid: false,
		},
		{
			name:        "Invalid - contains letters",
			phone:       "+123456789a",
			expectValid: false,
		},
		{
			name:        "Invalid - contains special characters",
			phone:       "+123-456-7890",
			expectValid: false,
		},
		{
			name:        "Invalid - empty string",
			phone:       "",
			expectValid: false,
		},
		{
			name:        "Invalid - only + symbol",
			phone:       "+",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phoneTest := PhoneTest{Phone: tt.phone}
			err := ValidateStruct(phoneTest)

			if tt.expectValid && err != nil {
				t.Errorf("Expected valid phone number %s, but got error: %v", tt.phone, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid phone number %s, but validation passed", tt.phone)
			}
		})
	}
}

func TestStrongPasswordValidator(t *testing.T) {
	RegisterCustomValidations()

	tests := []struct {
		name        string
		password    string
		expectValid bool
	}{
		{
			name:        "Valid strong password",
			password:    "MyP@ssw0rd123",
			expectValid: true,
		},
		{
			name:        "Valid strong password - minimal requirements",
			password:    "Aa1!bcde",
			expectValid: true,
		},
		{
			name:        "Valid strong password - with various special chars",
			password:    "Test123@#$%",
			expectValid: true,
		},
		{
			name:        "Invalid - too short (7 chars)",
			password:    "Aa1!bcd",
			expectValid: false,
		},
		{
			name:        "Invalid - no uppercase",
			password:    "myp@ssw0rd123",
			expectValid: false,
		},
		{
			name:        "Invalid - no lowercase",
			password:    "MYP@SSW0RD123",
			expectValid: false,
		},
		{
			name:        "Invalid - no digits",
			password:    "MyP@ssword!",
			expectValid: false,
		},
		{
			name:        "Invalid - no special characters",
			password:    "MyPassword123",
			expectValid: false,
		},
		{
			name:        "Invalid - only lowercase",
			password:    "password",
			expectValid: false,
		},
		{
			name:        "Invalid - only uppercase",
			password:    "PASSWORD",
			expectValid: false,
		},
		{
			name:        "Invalid - only digits",
			password:    "12345678",
			expectValid: false,
		},
		{
			name:        "Invalid - only special characters",
			password:    "!@#$%^&*",
			expectValid: false,
		},
		{
			name:        "Invalid - empty string",
			password:    "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordTest := PasswordTest{Password: tt.password}
			err := ValidateStruct(passwordTest)

			if tt.expectValid && err != nil {
				t.Errorf("Expected valid password %s, but got error: %v", tt.password, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid password %s, but validation passed", tt.password)
			}
		})
	}
}

// Test combined validation
func TestCombinedValidation(t *testing.T) {
	RegisterCustomValidations()

	type UserRegistration struct {
		AccountNumber string `validate:"required,account_number"`
		Phone         string `validate:"required,phone_th"`
		Password      string `validate:"required,strong_password"`
		Email         string `validate:"required,email"`
	}

	tests := []struct {
		name        string
		user        UserRegistration
		expectValid bool
	}{
		{
			name: "Valid user registration",
			user: UserRegistration{
				AccountNumber: "123456789012",
				Phone:         "+66812345678",
				Password:      "MyP@ssw0rd123",
				Email:         "user@example.com",
			},
			expectValid: true,
		},
		{
			name: "Invalid - bad account number",
			user: UserRegistration{
				AccountNumber: "12345",
				Phone:         "+66812345678",
				Password:      "MyP@ssw0rd123",
				Email:         "user@example.com",
			},
			expectValid: false,
		},
		{
			name: "Invalid - bad phone",
			user: UserRegistration{
				AccountNumber: "123456789012",
				Phone:         "66812345678",
				Password:      "MyP@ssw0rd123",
				Email:         "user@example.com",
			},
			expectValid: false,
		},
		{
			name: "Invalid - weak password",
			user: UserRegistration{
				AccountNumber: "123456789012",
				Phone:         "+66812345678",
				Password:      "password",
				Email:         "user@example.com",
			},
			expectValid: false,
		},
		{
			name: "Invalid - bad email",
			user: UserRegistration{
				AccountNumber: "123456789012",
				Phone:         "+66812345678",
				Password:      "MyP@ssw0rd123",
				Email:         "invalid-email",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.user)

			if tt.expectValid && err != nil {
				t.Errorf("Expected valid user registration, but got error: %v", err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("Expected invalid user registration, but validation passed")
			}
		})
	}
}

// Test custom error messages
func TestCustomValidatorErrorMessages(t *testing.T) {
	RegisterCustomValidations()

	tests := []struct {
		name            string
		testStruct      interface{}
		expectedMessage string
	}{
		{
			name:            "Account number validation error message",
			testStruct:      AccountTest{AccountNumber: "123"},
			expectedMessage: "AccountNumber must be exactly 12 digits",
		},
		{
			name:            "Phone validation error message",
			testStruct:      PhoneTest{Phone: "invalid-phone"},
			expectedMessage: "Phone must be a valid Thai phone number (format: +66xxxxxxxxx, 10-12 digits after +)",
		},
		{
			name:            "Strong password validation error message",
			testStruct:      PasswordTest{Password: "weak"},
			expectedMessage: "Password must be at least 8 characters long and contain uppercase, lowercase, digit, and special character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.testStruct)

			if err == nil {
				t.Errorf("Expected validation error, but got nil")
				return
			}

			// Convert error to our custom error type to access the details
			if validationErr, ok := err.(*response.ErrorResponse); ok {
				if details, ok := validationErr.Details.(map[string]interface{}); ok {
					if errors, ok := details["errors"].([]string); ok && len(errors) > 0 {
						if errors[0] != tt.expectedMessage {
							t.Errorf("Expected error message: %s, but got: %s", tt.expectedMessage, errors[0])
						}
					} else {
						t.Errorf("Expected error messages array, but got: %v", details["errors"])
					}
				} else {
					t.Errorf("Expected details map, but got: %v", validationErr.Details)
				}
			} else {
				t.Errorf("Expected ErrorResponse type, but got: %T", err)
			}
		})
	}
}

// Test multiple validation errors at once
func TestMultipleValidationErrors(t *testing.T) {
	RegisterCustomValidations()

	type MultiValidationTest struct {
		AccountNumber string `validate:"required,account_number"`
		Phone         string `validate:"required,phone_th"`
		Password      string `validate:"required,strong_password"`
		Email         string `validate:"required,email"`
	}

	invalidData := MultiValidationTest{
		AccountNumber: "123",          // Invalid: too short
		Phone:         "invalid",      // Invalid: wrong format
		Password:      "weak",         // Invalid: not strong enough
		Email:         "not-an-email", // Invalid: not email format
	}

	err := ValidateStruct(invalidData)

	if err == nil {
		t.Errorf("Expected validation errors, but got nil")
		return
	}

	// Convert error to our custom error type
	if validationErr, ok := err.(*response.ErrorResponse); ok {
		if details, ok := validationErr.Details.(map[string]interface{}); ok {
			if errors, ok := details["errors"].([]string); ok {
				// Should have 4 validation errors
				if len(errors) != 4 {
					t.Errorf("Expected 4 validation errors, but got %d: %v", len(errors), errors)
				}

				// Check that our custom error messages are included
				expectedMessages := []string{
					"AccountNumber must be exactly 12 digits",
					"Phone must be a valid Thai phone number (format: +66xxxxxxxxx, 10-12 digits after +)",
					"Password must be at least 8 characters long and contain uppercase, lowercase, digit, and special character",
					"Email must be a valid email address",
				}

				for _, expectedMsg := range expectedMessages {
					found := false
					for _, actualMsg := range errors {
						if actualMsg == expectedMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error message not found: %s\nActual messages: %v", expectedMsg, errors)
					}
				}
			}
		}
	}
}
