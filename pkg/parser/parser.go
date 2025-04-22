package parser

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

type SwiftCode struct {
	CountryISO2          string
	SwiftCode            string
	BankName             string
	Address              *string
	TownName             string
	CountryName          string
	Timezone             string
	IsHeadquarter        bool
	HeadquarterSWIFTCode *string
}

func ParseCSV(filePath string) ([]SwiftCode, []SwiftCode, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		log.Println("Error reading file header:", err)
		return nil, nil, err
	}

	var headquarters []SwiftCode
	var branches []SwiftCode

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		swiftCode := record[1]
		isHeadquarter := strings.HasSuffix(swiftCode, "XXX")
		var headquarterSWIFTCode *string
		if !isHeadquarter {
			hq := swiftCode[:8] + "XXX"
			headquarterSWIFTCode = &hq
		}

		address := record[4]
		var addressPtr *string
		if address != "" {
			addressPtr = &address
		}

		code := SwiftCode{
			CountryISO2:          strings.ToUpper(record[0]),
			SwiftCode:            swiftCode,
			BankName:             record[3],
			Address:              addressPtr,
			TownName:             record[5],
			CountryName:          record[6],
			Timezone:             record[7],
			IsHeadquarter:        isHeadquarter,
			HeadquarterSWIFTCode: headquarterSWIFTCode,
		}

		if isHeadquarter {
			headquarters = append(headquarters, code)
		} else {
			branches = append(branches, code)
		}
	}

	return headquarters, branches, nil
}
