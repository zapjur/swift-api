package models

type SwiftCode struct {
	CountryISO2          string      `json:"countryISO2"`
	SwiftCode            string      `json:"swiftCode"`
	BankName             string      `json:"bankName"`
	Address              *string     `json:"address,omitempty"`
	TownName             string      `json:"townName"`
	CountryName          string      `json:"countryName"`
	Timezone             string      `json:"timezone"`
	IsHeadquarter        bool        `json:"isHeadquarter"`
	HeadquarterSWIFTCode *string     `json:"headquarterSwiftCode,omitempty"`
	Branches             []SwiftCode `json:"branches,omitempty"`
}
