package validators

import (
	"testing"
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
