package handlers

import (
	"fmt"
	"strings"
)

func (r *CreateSwiftCodeRequest) Validate() error {

	if r.BankName == "" {
		return fmt.Errorf("bankName is required")
	}
	if r.CountryISO2 == "" || len(r.CountryISO2) != 2 {
		return fmt.Errorf("countryISO2 must be 2-letter code")
	}
	if r.CountryName == "" {
		return fmt.Errorf("countryName is required")
	}

	if r.Address == "" {
		return fmt.Errorf("address is required")
	}

	if r.SwiftCode == "" {
		return fmt.Errorf("swiftCode is required")
	}

	if len(r.SwiftCode) != 11 {
		return fmt.Errorf("swiftCode must be exactly 11 characters")
	}

	if r.IsHeadquarter && !strings.HasSuffix(r.SwiftCode, "XXX") {
		return fmt.Errorf("headquarter swiftCode must end with 'XXX'")
	}

	if !r.IsHeadquarter && strings.HasSuffix(r.SwiftCode, "XXX") {
		return fmt.Errorf("branch swiftCode cannot end with 'XXX'")
	}

	return nil
}
