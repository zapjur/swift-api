package repository

import (
	"database/sql"
	"log"
	"swift-api/pkg/models"
)

type Repository interface {
	InsertSwiftCodes(swiftCodes []models.SwiftCode) error
	GetSwiftCodeDetails(swiftCode string) (*models.SwiftCode, error)
	GetBranchesByHeadquarter(headquarterSWIFTCode string) ([]models.SwiftCode, error)
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
