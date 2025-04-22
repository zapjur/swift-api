package parser

import (
	"log"
)

func FillMissingHeadquarters(hq []SwiftCode, branches []SwiftCode) []SwiftCode {
	hqMap := make(map[string]struct{})
	for _, h := range hq {
		hqMap[h.SwiftCode] = struct{}{}
	}

	var placeholderHQs []SwiftCode

	for _, b := range branches {
		if b.HeadquarterSWIFTCode == nil {
			continue
		}
		hqCode := *b.HeadquarterSWIFTCode
		if _, exists := hqMap[hqCode]; !exists {
			log.Printf("Adding placeholder HQ: %s", hqCode)
			placeholderHQs = append(placeholderHQs, SwiftCode{
				CountryISO2:          "ZZ",
				SwiftCode:            hqCode,
				BankName:             "UNKNOWN",
				Address:              nil,
				TownName:             "UNKNOWN",
				CountryName:          "UNKNOWN",
				Timezone:             "Etc/UTC",
				IsHeadquarter:        true,
				HeadquarterSWIFTCode: nil,
			})
			hqMap[hqCode] = struct{}{}
		}
	}

	return append(hq, placeholderHQs...)
}
