package repository

import (
	"database/sql"
	"log"
	"strings"
	"swift-api/pkg/models"
)

type Repository interface {
	InsertSwiftCodes(swiftCodes []models.SwiftCode) error
	GetSwiftCodeDetails(swiftCode string) (*models.SwiftCode, error)
	GetBranchesByHeadquarter(headquarterSWIFTCode string) ([]models.SwiftCode, error)
	GetSwiftCodesByCountry(iso2 string) ([]models.SwiftCode, string, error)
	HeadquarterExists(swiftCode string) (bool, error)
	SwiftCodeExists(swiftCode string) (bool, error)
	IsPlaceholder(swiftCode string) (bool, error)
	UpdatePlaceholderSwiftCode(code models.SwiftCode) error
	DeleteSwiftCode(swiftCode string) error
}

type Repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &Repo{db: db}
}

func (r *Repo) InsertSwiftCodes(swiftCodes []models.SwiftCode) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	for _, code := range swiftCodes {

		var address any
		if code.Address == nil {
			address = nil
		} else {
			address = *code.Address
		}

		var hqCode any
		if code.HeadquarterSWIFTCode == nil {
			hqCode = nil
		} else {
			hqCode = *code.HeadquarterSWIFTCode
		}

		_, err := tx.Exec(`
			INSERT INTO swift_codes (swift_code, bank_name, address, town_name, country_iso2, country_name, timezone, is_headquarter, headquarter_swift_code)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (swift_code) DO NOTHING`,
			code.SwiftCode, code.BankName, address, code.TownName, code.CountryISO2, code.CountryName, code.Timezone, code.IsHeadquarter, hqCode)

		if err != nil {
			log.Println("Error inserting data:", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	log.Println("Swift codes inserted successfully")
	return nil
}

func (r *Repo) GetSwiftCodeDetails(swiftCode string) (*models.SwiftCode, error) {
	row := r.db.QueryRow(`
		SELECT swift_code, bank_name, address, town_name, country_iso2, country_name, timezone, is_headquarter, headquarter_swift_code
		FROM swift_codes WHERE swift_code = $1`, swiftCode)

	var code models.SwiftCode
	err := row.Scan(&code.SwiftCode, &code.BankName, &code.Address, &code.TownName, &code.CountryISO2, &code.CountryName, &code.Timezone, &code.IsHeadquarter, &code.HeadquarterSWIFTCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Println("Error fetching SWIFT code details:", err)
		return nil, err
	}
	return &code, nil
}

func (r *Repo) GetBranchesByHeadquarter(headquarterSWIFTCode string) ([]models.SwiftCode, error) {
	rows, err := r.db.Query(`
		SELECT swift_code, bank_name, address, town_name, country_iso2, country_name, timezone, is_headquarter, headquarter_swift_code
		FROM swift_codes WHERE headquarter_swift_code = $1`, headquarterSWIFTCode)
	if err != nil {
		log.Println("Error fetching branches:", err)
		return nil, err
	}
	defer rows.Close()

	var branches []models.SwiftCode
	for rows.Next() {
		var branch models.SwiftCode
		err := rows.Scan(&branch.SwiftCode, &branch.BankName, &branch.Address, &branch.TownName, &branch.CountryISO2, &branch.CountryName, &branch.Timezone, &branch.IsHeadquarter, &branch.HeadquarterSWIFTCode)
		if err != nil {
			log.Println("Error scanning branch:", err)
			return nil, err
		}
		branches = append(branches, branch)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error with rows:", err)
		return nil, err
	}

	return branches, nil
}

func (r *Repo) GetSwiftCodesByCountry(iso2 string) ([]models.SwiftCode, string, error) {
	rows, err := r.db.Query(`
		SELECT swift_code, bank_name, address, town_name, country_iso2, country_name, timezone, is_headquarter, headquarter_swift_code
		FROM swift_codes WHERE country_iso2 = $1`, strings.ToUpper(iso2))
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var codes []models.SwiftCode
	var countryName string

	for rows.Next() {
		var c models.SwiftCode
		err = rows.Scan(&c.SwiftCode, &c.BankName, &c.Address, &c.TownName, &c.CountryISO2, &c.CountryName, &c.Timezone, &c.IsHeadquarter, &c.HeadquarterSWIFTCode)
		if err != nil {
			return nil, "", err
		}
		codes = append(codes, c)
		countryName = c.CountryName
	}

	return codes, countryName, nil
}

func (r *Repo) HeadquarterExists(swiftCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM swift_codes 
			WHERE swift_code = $1 AND is_headquarter = TRUE
		)
	`, swiftCode).Scan(&exists)

	if err != nil {
		log.Println("Error checking headquarter existence:", err)
		return false, err
	}

	return exists, nil
}

func (r *Repo) SwiftCodeExists(swiftCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM swift_codes WHERE swift_code = $1)
	`, swiftCode).Scan(&exists)

	if err != nil {
		log.Println("Error checking SWIFT code existence:", err)
		return false, err
	}

	return exists, nil
}

func (r *Repo) IsPlaceholder(swiftCode string) (bool, error) {
	var bankName, timezone string
	err := r.db.QueryRow(`
		SELECT bank_name, timezone FROM swift_codes WHERE swift_code = $1
	`, swiftCode).Scan(&bankName, &timezone)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return bankName == "UNKNOWN" && timezone == "Etc/UTC", nil
}

func (r *Repo) UpdatePlaceholderSwiftCode(code models.SwiftCode) error {
	_, err := r.db.Exec(`
		UPDATE swift_codes SET
			bank_name = $1,
			address = $2,
			town_name = $3,
			country_iso2 = $4,
			country_name = $5,
			timezone = $6,
			is_headquarter = $7,
			headquarter_swift_code = $8
		WHERE swift_code = $9
		AND bank_name = 'UNKNOWN' AND timezone = 'Etc/UTC'
	`, code.BankName, code.Address, code.TownName, code.CountryISO2,
		code.CountryName, code.Timezone, code.IsHeadquarter, code.HeadquarterSWIFTCode, code.SwiftCode)

	return err
}

func (r *Repo) DeleteSwiftCode(swiftCode string) error {
	_, err := r.db.Exec(`DELETE FROM swift_codes WHERE swift_code = $1`, swiftCode)
	return err
}
