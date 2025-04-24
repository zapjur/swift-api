package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"swift-api/pkg/handlers"
	"testing"
)

func TestValidateSwiftCodeRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   handlers.CreateSwiftCodeRequest
		expects string
	}{
		{
			name:    "Missing bankName",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "ABCDEF12XXX", CountryISO2: "PL", CountryName: "Poland", Address: "Main St", IsHeadquarter: true},
			expects: "bankName is required",
		},
		{
			name:    "Invalid ISO2 code",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "ABCDEF12XXX", BankName: "Bank", CountryISO2: "P", CountryName: "Poland", Address: "Main St", IsHeadquarter: true},
			expects: "countryISO2 must be 2-letter code",
		},
		{
			name:    "Missing country name",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "ABCDEF12XXX", BankName: "Bank", CountryISO2: "PL", Address: "Main St", IsHeadquarter: true},
			expects: "countryName is required",
		},
		{
			name:    "Missing address",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "ABCDEF12XXX", BankName: "Bank", CountryISO2: "PL", CountryName: "Poland", IsHeadquarter: true},
			expects: "address is required",
		},
		{
			name:    "Missing swiftCode",
			input:   handlers.CreateSwiftCodeRequest{BankName: "Bank", CountryISO2: "PL", CountryName: "Poland", Address: "Main St", IsHeadquarter: true},
			expects: "swiftCode is required",
		},
		{
			name:    "Too short swiftCode",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "TOOSHORT", BankName: "Bank", CountryISO2: "PL", CountryName: "Poland", Address: "Main St", IsHeadquarter: true},
			expects: "swiftCode must be exactly 11 characters",
		},
		{
			name:    "HQ without XXX",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "ABCDEF12345", BankName: "Bank", CountryISO2: "PL", CountryName: "Poland", Address: "Main St", IsHeadquarter: true},
			expects: "headquarter swiftCode must end with 'XXX'",
		},
		{
			name:    "Branch ends with XXX",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "ABCDEF12XXX", BankName: "Bank", CountryISO2: "PL", CountryName: "Poland", Address: "Main St", IsHeadquarter: false},
			expects: "branch swiftCode cannot end with 'XXX'",
		},
		{
			name:    "Valid HQ",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "BANKPLPWXXX", BankName: "Bank", CountryISO2: "PL", CountryName: "Poland", Address: "HQ Address", IsHeadquarter: true},
			expects: "",
		},
		{
			name:    "Valid Branch",
			input:   handlers.CreateSwiftCodeRequest{SwiftCode: "BANKPLPW001", BankName: "Branch", CountryISO2: "PL", CountryName: "Poland", Address: "Branch Address", IsHeadquarter: false},
			expects: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.expects == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expects)
			}
		})
	}
}
