package repository_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"swift-api/pkg/models"
	"swift-api/pkg/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	return db
}

func TestInsertSwiftCodes(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	swiftCode := models.SwiftCode{
		SwiftCode:     "TESTINSERT1",
		BankName:      "Test Bank",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		TownName:      "WARSAW",
		Timezone:      "Europe/Warsaw",
		IsHeadquarter: true,
		Address:       nil,
	}

	err := repo.InsertSwiftCodes([]models.SwiftCode{swiftCode})
	assert.NoError(t, err)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM swift_codes WHERE swift_code = 'TESTINSERT1'`).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestInsertSwiftCodes_AdvancedCases(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	t.Run("Insert new HQ", func(t *testing.T) {
		err := repo.InsertSwiftCodes([]models.SwiftCode{
			{
				SwiftCode:     "NEWPLLHQXXX",
				BankName:      "New HQ Bank",
				CountryISO2:   "PL",
				CountryName:   "POLAND",
				TownName:      "WARSAW",
				IsHeadquarter: true,
				Address:       strPtr("HQ Main St"),
				Timezone:      "Europe/Warsaw",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("Insert branch with existing HQ", func(t *testing.T) {
		err := repo.InsertSwiftCodes([]models.SwiftCode{
			{
				SwiftCode:            "NEWPLLHQ001",
				BankName:             "Branch Bank",
				CountryISO2:          "PL",
				CountryName:          "POLAND",
				TownName:             "KRAKOW",
				IsHeadquarter:        false,
				Address:              strPtr("Branch St"),
				Timezone:             "Europe/Warsaw",
				HeadquarterSWIFTCode: strPtr("NEWPLLHQXXX"),
			},
		})
		assert.NoError(t, err)
	})

	t.Run("Duplicate HQ", func(t *testing.T) {
		err := repo.InsertSwiftCodes([]models.SwiftCode{
			{
				SwiftCode:     "NEWPLLHQXXX",
				BankName:      "Duplicate HQ",
				CountryISO2:   "PL",
				CountryName:   "POLAND",
				TownName:      "WARSAW",
				IsHeadquarter: true,
				Address:       strPtr("Some address"),
				Timezone:      "Europe/Warsaw",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("Insert branch with missing HQ (should fail on FK)", func(t *testing.T) {
		err := repo.InsertSwiftCodes([]models.SwiftCode{
			{
				SwiftCode:            "MISSINGHQ001",
				BankName:             "Orphan Branch",
				CountryISO2:          "PL",
				CountryName:          "POLAND",
				TownName:             "GDANSK",
				IsHeadquarter:        false,
				Address:              strPtr("Nowhere St"),
				Timezone:             "Europe/Warsaw",
				HeadquarterSWIFTCode: strPtr("DOESNOTEXISTXXX"),
			},
		})
		assert.Error(t, err)
	})

	t.Run("Insert HQ with nil address", func(t *testing.T) {
		err := repo.InsertSwiftCodes([]models.SwiftCode{
			{
				SwiftCode:     "NULLADDRXXX",
				BankName:      "No Address HQ",
				CountryISO2:   "PL",
				CountryName:   "POLAND",
				TownName:      "LODZ",
				IsHeadquarter: true,
				Address:       nil,
				Timezone:      "Europe/Warsaw",
			},
		})
		assert.NoError(t, err)
	})
}

func strPtr(s string) *string {
	return &s
}

func TestGetSwiftCodeDetails(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	hq := models.SwiftCode{
		SwiftCode:     "DETATESTXXX",
		BankName:      "Detail Test HQ",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		TownName:      "WARSZAWA",
		IsHeadquarter: true,
		Address:       strPtr("Detail St 1"),
		Timezone:      "Europe/Warsaw",
	}
	err := repo.InsertSwiftCodes([]models.SwiftCode{hq})
	assert.NoError(t, err)

	t.Run("Get existing SWIFT code", func(t *testing.T) {
		result, err := repo.GetSwiftCodeDetails("DETATESTXXX")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "DETATESTXXX", result.SwiftCode)
		assert.Equal(t, "Detail Test HQ", result.BankName)
		assert.Equal(t, "WARSZAWA", result.TownName)
		assert.True(t, result.IsHeadquarter)
		assert.NotNil(t, result.Address)
	})

	t.Run("Get non-existent SWIFT code", func(t *testing.T) {
		result, err := repo.GetSwiftCodeDetails("DOESNOTEXIS")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestGetBranchesByHeadquarter(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	hqCode := "BRANCHTEXXX"

	err := repo.InsertSwiftCodes([]models.SwiftCode{
		{
			SwiftCode:     hqCode,
			BankName:      "Test HQ",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "WARSZAWA",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		},
		{
			SwiftCode:            "BRANCHTE001",
			BankName:             "Branch One",
			CountryISO2:          "PL",
			CountryName:          "POLAND",
			TownName:             "WARSZAWA",
			IsHeadquarter:        false,
			Timezone:             "Europe/Warsaw",
			HeadquarterSWIFTCode: &hqCode,
		},
		{
			SwiftCode:            "BRANCHTE002",
			BankName:             "Branch Two",
			CountryISO2:          "PL",
			CountryName:          "POLAND",
			TownName:             "POZNAN",
			IsHeadquarter:        false,
			Timezone:             "Europe/Warsaw",
			HeadquarterSWIFTCode: &hqCode,
		},
		{
			SwiftCode:     "NOCHILDSXXX",
			BankName:      "Lonely HQ",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "KATOWICE",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		},
	})
	assert.NoError(t, err)

	t.Run("Get branches for existing HQ", func(t *testing.T) {
		branches, err := repo.GetBranchesByHeadquarter(hqCode)
		assert.NoError(t, err)
		assert.Len(t, branches, 2)

		var codes []string
		for _, b := range branches {
			codes = append(codes, b.SwiftCode)
		}
		assert.Contains(t, codes, "BRANCHTE001")
		assert.Contains(t, codes, "BRANCHTE002")
	})

	t.Run("Get branches for HQ with no branches", func(t *testing.T) {
		branches, err := repo.GetBranchesByHeadquarter("NOCHILDSXXX")
		assert.NoError(t, err)
		assert.Len(t, branches, 0)
	})

	t.Run("Get branches for non-existent HQ", func(t *testing.T) {
		branches, err := repo.GetBranchesByHeadquarter("UNKNOWNNXXX")
		assert.NoError(t, err)
		assert.Len(t, branches, 0)
	})
}

func TestGetSwiftCodesByCountry(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	err := repo.InsertSwiftCodes([]models.SwiftCode{
		{
			SwiftCode:     "PLCOUNTRXXX",
			BankName:      "Polish HQ",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "WARSAW",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		},
		{
			SwiftCode:            "PLCOUNTR001",
			BankName:             "Polish Branch",
			CountryISO2:          "PL",
			CountryName:          "POLAND",
			TownName:             "KRAKOW",
			IsHeadquarter:        false,
			Timezone:             "Europe/Warsaw",
			HeadquarterSWIFTCode: strPtr("PLCOUNTRXXX"),
		},
		{
			SwiftCode:     "USCOUNTRXXX",
			BankName:      "US HQ",
			CountryISO2:   "US",
			CountryName:   "UNITED STATES",
			TownName:      "NEW YORK",
			IsHeadquarter: true,
			Timezone:      "America/New_York",
		},
	})
	assert.NoError(t, err)

	t.Run("Get codes for PL", func(t *testing.T) {
		codes, countryName, err := repo.GetSwiftCodesByCountry("pl")
		assert.NoError(t, err)
		assert.Len(t, codes, 2)
		assert.Equal(t, "POLAND", countryName)

		var swiftCodes []string
		for _, c := range codes {
			swiftCodes = append(swiftCodes, c.SwiftCode)
		}
		assert.Contains(t, swiftCodes, "PLCOUNTRXXX")
		assert.Contains(t, swiftCodes, "PLCOUNTR001")
	})

	t.Run("Get codes for US", func(t *testing.T) {
		codes, countryName, err := repo.GetSwiftCodesByCountry("us")
		assert.NoError(t, err)
		assert.Len(t, codes, 1)
		assert.Equal(t, "UNITED STATES", countryName)
		assert.Equal(t, "USCOUNTRXXX", codes[0].SwiftCode)
	})

	t.Run("No results for XX", func(t *testing.T) {
		codes, countryName, err := repo.GetSwiftCodesByCountry("xx")
		assert.NoError(t, err)
		assert.Empty(t, codes)
		assert.Equal(t, "", countryName)
	})
}

func TestHeadquarterExists(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	hqCode := "EXISTHQQXXX"
	nonHqCode := "NOTAHQQQ001"
	missingCode := "DOESNOTEXIS"

	err := repo.InsertSwiftCodes([]models.SwiftCode{
		{
			SwiftCode:     hqCode,
			BankName:      "Test HQ",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "WARSAW",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		},
		{
			SwiftCode:     nonHqCode,
			BankName:      "Just a branch",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "KRAKOW",
			IsHeadquarter: false,
			Timezone:      "Europe/Warsaw",
		},
	})
	assert.NoError(t, err)

	t.Run("Existing HQ returns true", func(t *testing.T) {
		exists, err := repo.HeadquarterExists(hqCode)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Existing non-HQ returns false", func(t *testing.T) {
		exists, err := repo.HeadquarterExists(nonHqCode)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Non-existent code returns false", func(t *testing.T) {
		exists, err := repo.HeadquarterExists(missingCode)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestSwiftCodeExists(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	existingCode := "EXISTSCC001"
	nonExistingCode := "NOSUCHCODEE"

	err := repo.InsertSwiftCodes([]models.SwiftCode{
		{
			SwiftCode:     existingCode,
			BankName:      "Some Bank",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "WARSZAWA",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		},
	})
	assert.NoError(t, err)

	t.Run("Existing SWIFT code returns true", func(t *testing.T) {
		exists, err := repo.SwiftCodeExists(existingCode)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Non-existing SWIFT code returns false", func(t *testing.T) {
		exists, err := repo.SwiftCodeExists(nonExistingCode)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestIsPlaceholder(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	placeholderCode := "PLACEHOLXXX"
	normalCode := "NORMALHQXXX"

	err := repo.InsertSwiftCodes([]models.SwiftCode{
		{
			SwiftCode:     placeholderCode,
			BankName:      "UNKNOWN",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "UNKNOWN",
			IsHeadquarter: true,
			Timezone:      "Etc/UTC",
		},
		{
			SwiftCode:     normalCode,
			BankName:      "Normal HQ",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "WARSZAWA",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		},
	})
	assert.NoError(t, err)

	t.Run("Recognize placeholder HQ", func(t *testing.T) {
		isPlaceholder, err := repo.IsPlaceholder(placeholderCode)
		assert.NoError(t, err)
		assert.True(t, isPlaceholder)
	})

	t.Run("Recognize non-placeholder HQ", func(t *testing.T) {
		isPlaceholder, err := repo.IsPlaceholder(normalCode)
		assert.NoError(t, err)
		assert.False(t, isPlaceholder)
	})

	t.Run("Non-existent code returns false", func(t *testing.T) {
		isPlaceholder, err := repo.IsPlaceholder("NOPEEEEEXXX")
		assert.NoError(t, err)
		assert.False(t, isPlaceholder)
	})
}

func TestUpdatePlaceholderSwiftCode(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	t.Run("Update existing placeholder with real data", func(t *testing.T) {
		placeholderCode := "PLACEHOLXXX"

		err := repo.InsertSwiftCodes([]models.SwiftCode{
			{
				SwiftCode:     placeholderCode,
				BankName:      "UNKNOWN",
				CountryISO2:   "ZZ",
				CountryName:   "UNKNOWN",
				TownName:      "UNKNOWN",
				IsHeadquarter: true,
				Timezone:      "Etc/UTC",
				Address:       nil,
			},
		})
		assert.NoError(t, err)

		fullCode := models.SwiftCode{
			SwiftCode:     placeholderCode,
			BankName:      "Updated Bank",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "WARSZAWA",
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
			Address:       strPtr("HQ Updated Address"),
		}

		err = repo.UpdatePlaceholderSwiftCode(fullCode)
		assert.NoError(t, err)

		updated, err := repo.GetSwiftCodeDetails(placeholderCode)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "Updated Bank", updated.BankName)
		assert.Equal(t, "PL", updated.CountryISO2)
		assert.Equal(t, "WARSZAWA", updated.TownName)
		assert.Equal(t, "Europe/Warsaw", updated.Timezone)
		assert.NotNil(t, updated.Address)
		assert.Equal(t, "HQ Updated Address", *updated.Address)
	})

	t.Run("Update non-existing placeholder", func(t *testing.T) {
		code := models.SwiftCode{
			SwiftCode:     "NOEXISTSXXX",
			BankName:      "Real HQ Bank",
			CountryISO2:   "PL",
			CountryName:   "POLAND",
			TownName:      "LUBLIN",
			Address:       strPtr("Some St 7"),
			IsHeadquarter: true,
			Timezone:      "Europe/Warsaw",
		}

		err := repo.UpdatePlaceholderSwiftCode(code)
		assert.NoError(t, err)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM swift_codes WHERE swift_code = $1`, "NOEXISTSXXX").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestDeleteSwiftCode(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)

	defer db.Exec("DELETE FROM swift_codes")

	code := models.SwiftCode{
		SwiftCode:     "DELTESTTXXX",
		BankName:      "Delete Test Bank",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		TownName:      "WARSZAWA",
		IsHeadquarter: true,
		Timezone:      "Europe/Warsaw",
	}

	err := repo.InsertSwiftCodes([]models.SwiftCode{code})
	assert.NoError(t, err)

	err = repo.DeleteSwiftCode("DELTESTTXXX")
	assert.NoError(t, err)

	exists, err := repo.SwiftCodeExists("DELTESTTXXX")
	assert.NoError(t, err)
	assert.False(t, exists)

	t.Run("Delete non-existent SWIFT code", func(t *testing.T) {
		err := repo.DeleteSwiftCode("DOESNOTEXIS")
		assert.NoError(t, err)
	})
}
