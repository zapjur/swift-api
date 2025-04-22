package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"swift-api/pkg/repository"
)

type Handler struct {
	repo repository.Repository
}

type BranchResponse struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	CountryName   string `json:"countryName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

type HeadquarterResponse struct {
	Address       string               `json:"address"`
	BankName      string               `json:"bankName"`
	CountryISO2   string               `json:"countryISO2"`
	CountryName   string               `json:"countryName"`
	IsHeadquarter bool                 `json:"isHeadquarter"`
	SwiftCode     string               `json:"swiftCode"`
	Branches      []BranchInHQResponse `json:"branches"`
}

type BranchInHQResponse struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

type CountryResponse struct {
	CountryISO2 string               `json:"countryISO2"`
	CountryName string               `json:"countryName"`
	SwiftCodes  []BranchInHQResponse `json:"swiftCodes"`
}

func NewHandler(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetSwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]

	code, err := h.repo.GetSwiftCodeDetails(swiftCode)
	if err != nil {
		http.Error(w, "Error retrieving SWIFT code", http.StatusInternalServerError)
		return
	}
	if code == nil {
		http.Error(w, "SWIFT code not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if code.IsHeadquarter {
		branches, err := h.repo.GetBranchesByHeadquarter(swiftCode)
		if err != nil {
			http.Error(w, "Error retrieving branches", http.StatusInternalServerError)
			return
		}
		var branchResponses []BranchInHQResponse
		for _, b := range branches {
			branchResponses = append(branchResponses, BranchInHQResponse{
				Address:       *b.Address,
				BankName:      b.BankName,
				CountryISO2:   b.CountryISO2,
				IsHeadquarter: b.IsHeadquarter,
				SwiftCode:     b.SwiftCode,
			})
		}

		resp := HeadquarterResponse{
			Address:       *code.Address,
			BankName:      code.BankName,
			CountryISO2:   code.CountryISO2,
			CountryName:   code.CountryName,
			IsHeadquarter: true,
			SwiftCode:     code.SwiftCode,
			Branches:      branchResponses,
		}

		if err = json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	resp := BranchResponse{
		Address:       *code.Address,
		BankName:      code.BankName,
		CountryISO2:   code.CountryISO2,
		CountryName:   code.CountryName,
		IsHeadquarter: false,
		SwiftCode:     code.SwiftCode,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *Handler) GetSwiftCodesByCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	iso2 := vars["countryISO2code"]

	codes, countryName, err := h.repo.GetSwiftCodesByCountry(iso2)
	if err != nil {
		http.Error(w, "Error retrieving SWIFT codes", http.StatusInternalServerError)
		return
	}
	if len(codes) == 0 {
		http.Error(w, "No SWIFT codes found for this country", http.StatusNotFound)
		return
	}

	var respCodes []BranchInHQResponse
	for _, code := range codes {
		respCodes = append(respCodes, BranchInHQResponse{
			Address:       *code.Address,
			BankName:      code.BankName,
			CountryISO2:   code.CountryISO2,
			IsHeadquarter: code.IsHeadquarter,
			SwiftCode:     code.SwiftCode,
		})

	}

	resp := CountryResponse{
		CountryISO2: strings.ToUpper(iso2),
		CountryName: countryName,
		SwiftCodes:  respCodes,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateSwiftCode(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) DeleteSwiftCode(w http.ResponseWriter, r *http.Request) {

}
