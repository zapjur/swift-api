package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"swift-api/pkg/models"
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
type SwiftCodeInCountryResponse = BranchInHQResponse

type CountryResponse struct {
	CountryISO2 string               `json:"countryISO2"`
	CountryName string               `json:"countryName"`
	SwiftCodes  []BranchInHQResponse `json:"swiftCodes"`
}

type CreateSwiftCodeRequest struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	CountryName   string `json:"countryName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
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
				Address:       strOrEmpty(b.Address),
				BankName:      b.BankName,
				CountryISO2:   b.CountryISO2,
				IsHeadquarter: b.IsHeadquarter,
				SwiftCode:     b.SwiftCode,
			})
		}

		resp := HeadquarterResponse{
			Address:       strOrEmpty(code.Address),
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
		Address:       strOrEmpty(code.Address),
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

	var respCodes []SwiftCodeInCountryResponse
	for _, code := range codes {
		respCodes = append(respCodes, SwiftCodeInCountryResponse{
			Address:       strOrEmpty(code.Address),
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
	var req CreateSwiftCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.CountryISO2 = strings.ToUpper(req.CountryISO2)
	req.CountryName = strings.ToUpper(req.CountryName)

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newCode := models.SwiftCode{
		Address:       &req.Address,
		BankName:      req.BankName,
		CountryISO2:   req.CountryISO2,
		CountryName:   req.CountryName,
		IsHeadquarter: req.IsHeadquarter,
		SwiftCode:     req.SwiftCode,
	}

	if !req.IsHeadquarter {
		hqCode := req.SwiftCode[:8] + "XXX"
		exists, err := h.repo.HeadquarterExists(hqCode)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		if !exists {
			placeholder := models.SwiftCode{
				SwiftCode:            hqCode,
				BankName:             "UNKNOWN",
				TownName:             "UNKNOWN",
				CountryISO2:          req.CountryISO2,
				CountryName:          req.CountryName,
				Timezone:             "Etc/UTC",
				IsHeadquarter:        true,
				Address:              nil,
				HeadquarterSWIFTCode: nil,
			}
			err = h.repo.InsertSwiftCodes([]models.SwiftCode{placeholder})
			if err != nil {
				http.Error(w, "Failed to insert placeholder HQ", http.StatusInternalServerError)
				return
			}
		}

		hqCodePtr := hqCode
		newCode.HeadquarterSWIFTCode = &hqCodePtr
	}

	exists, err := h.repo.SwiftCodeExists(newCode.SwiftCode)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	if exists {
		isPlaceholder, err := h.repo.IsPlaceholder(newCode.SwiftCode)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		if isPlaceholder {
			err = h.repo.UpdatePlaceholderSwiftCode(newCode)
			if err != nil {
				http.Error(w, "Failed to update placeholder SWIFT code", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "SWIFT code already exists", http.StatusConflict)
			return
		}
	} else {
		if err := h.repo.InsertSwiftCodes([]models.SwiftCode{newCode}); err != nil {
			http.Error(w, "Failed to insert SWIFT code", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{
		"message": "SWIFT code added successfully",
	})
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteSwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]

	if swiftCode == "" {
		http.Error(w, "SWIFT code is required", http.StatusBadRequest)
		return
	}

	if len(swiftCode) != 11 {
		http.Error(w, "SWIFT code must be exactly 11 characters", http.StatusBadRequest)
		return
	}

	if strings.HasSuffix(swiftCode, "XXX") {
		branches, err := h.repo.GetBranchesByHeadquarter(swiftCode)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		if len(branches) > 0 {
			http.Error(w, "Cannot delete headquarter with existing branches", http.StatusConflict)
			return
		}
	}

	err := h.repo.DeleteSwiftCode(swiftCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "SWIFT code not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error deleting SWIFT code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{
		"message": "SWIFT code deleted successfully",
	})
}
