package parser

import (
	"path/filepath"
	"swift-api/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCSV(t *testing.T) {
	path := filepath.Join("testdata", "test_swift_codes.csv")
	hq, branches, err := ParseCSV(path)

	assert.NoError(t, err)
	assert.Len(t, hq, 1)
	assert.Len(t, branches, 1)

	assert.Equal(t, "TESTPLHQXXX", hq[0].SwiftCode)
	assert.Equal(t, "Test Bank", hq[0].BankName)

	assert.Equal(t, "TESTPLHQ001", branches[0].SwiftCode)
	assert.Equal(t, "Test Branch", branches[0].BankName)
	assert.NotNil(t, branches[0].HeadquarterSWIFTCode)
	assert.Equal(t, "TESTPLHQXXX", *branches[0].HeadquarterSWIFTCode)
}

func TestParseCSV_AdvancedCases(t *testing.T) {
	path := filepath.Join("testdata", "test_swift_codes_advanced.csv")
	hq, branches, err := ParseCSV(path)

	assert.NoError(t, err)
	assert.Len(t, hq, 3)
	assert.Len(t, branches, 3)

	// Helper to find code by swiftCode
	findCode := func(codes []models.SwiftCode, swiftCode string) *models.SwiftCode {
		for _, c := range codes {
			if c.SwiftCode == swiftCode {
				return &c
			}
		}
		return nil
	}

	hq1 := findCode(hq, "TESTPLHQXXX")
	assert.NotNil(t, hq1)
	assert.Equal(t, "Test HQ Bank", hq1.BankName)

	hq2 := findCode(hq, "GERMDEFFXXX")
	assert.NotNil(t, hq2)
	assert.Equal(t, "German HQ", hq2.BankName)

	hq3 := findCode(hq, "BANKFRPPXXX")
	assert.NotNil(t, hq3)
	assert.Equal(t, "French HQ", hq3.BankName)
	assert.Nil(t, hq3.Address)

	br1 := findCode(branches, "TESTPLHQ001")
	assert.NotNil(t, br1)
	assert.Equal(t, "Test Branch 1", br1.BankName)
	assert.Equal(t, "TESTPLHQXXX", *br1.HeadquarterSWIFTCode)

	br2 := findCode(branches, "TESTPLHQ002")
	assert.NotNil(t, br2)
	assert.Equal(t, "Test Branch 2", br2.BankName)
	assert.Equal(t, "TESTPLHQXXX", *br2.HeadquarterSWIFTCode)

	br3 := findCode(branches, "GERMDEFF001")
	assert.NotNil(t, br3)
	assert.Equal(t, "German Branch", br3.BankName)
	assert.Equal(t, "GERMDEFFXXX", *br3.HeadquarterSWIFTCode)
}

func TestParseCSV_FileNotFound(t *testing.T) {
	_, _, err := ParseCSV("non_existent_file.csv")
	assert.Error(t, err)
}

func TestParseCSV_InvalidRecord(t *testing.T) {
	path := filepath.Join("testdata", "invalid_test_swift_codes.csv")
	hq, branches, err := ParseCSV(path)

	assert.NoError(t, err)
	assert.Len(t, hq, 0)
	assert.Len(t, branches, 0)
}

func TestParseCSV_CountryUppercase(t *testing.T) {
	path := filepath.Join("testdata", "uppercase_country_test_swift_codes.csv")
	hq, _, err := ParseCSV(path)

	assert.NoError(t, err)
	assert.Equal(t, "PL", hq[0].CountryISO2)
	assert.Equal(t, "POLAND", hq[0].CountryName)
}

func TestParseCSV_ExtraColumnsIgnored(t *testing.T) {
	path := filepath.Join("testdata", "extra_cols_test_swift_codes.csv")
	hq, branches, err := ParseCSV(path)

	assert.NoError(t, err)
	assert.Len(t, hq, 1)
	assert.Len(t, branches, 0)

	assert.Equal(t, "EXTRAPLXXX", hq[0].SwiftCode)
	assert.Equal(t, "Extra Bank", hq[0].BankName)
	assert.Equal(t, "POLAND", hq[0].CountryName)
}

func TestFillMissingHeadquarters(t *testing.T) {
	branch1 := models.SwiftCode{
		SwiftCode:            "BANKPLPW001",
		HeadquarterSWIFTCode: ptr("BANKPLPWXXX"),
	}
	branch2 := models.SwiftCode{
		SwiftCode:            "BANKPLPW002",
		HeadquarterSWIFTCode: ptr("BANKPLPWXXX"),
	}
	branch3 := models.SwiftCode{
		SwiftCode:            "BANKDEFF001",
		HeadquarterSWIFTCode: ptr("BANKDEFFXXX"),
	}

	existingHQ := models.SwiftCode{
		SwiftCode:     "BANKDEFFXXX",
		IsHeadquarter: true,
	}

	result := FillMissingHeadquarters([]models.SwiftCode{existingHQ}, []models.SwiftCode{branch1, branch2, branch3})

	var foundPlaceholder *models.SwiftCode
	for _, hq := range result {
		if hq.SwiftCode == "BANKPLPWXXX" {
			foundPlaceholder = &hq
			break
		}
	}

	assert.Len(t, result, 2) // 1 existing + 1 placeholder
	assert.NotNil(t, foundPlaceholder)
	assert.Equal(t, "UNKNOWN", foundPlaceholder.BankName)
	assert.Equal(t, "ZZ", foundPlaceholder.CountryISO2)
}

func ptr(s string) *string {
	return &s
}
