package repository

import (
	"database/sql"
	"log"
	"swift-api/pkg/parser"
)

type Repository interface {
	InsertSwiftCodes(swiftCodes []parser.SwiftCode) error
}

type Repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &Repo{db: db}
}

func (r *Repo) InsertSwiftCodes(swiftCodes []parser.SwiftCode) error {
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
