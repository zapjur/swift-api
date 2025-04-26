package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"swift-api/pkg/models"
	"swift-api/pkg/repository"
)

type Handler struct {
	Repo repository.Repository
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
	return &Handler{Repo: repo}
}

func (h *Handler) GetSwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]

	if swiftCode == "" {
		writeError(w, http.StatusBadRequest, "SWIFT code is required")
		return
	}

	if len(swiftCode) != 11 {
		writeError(w, http.StatusBadRequest, "SWIFT code must be exactly 11 characters")
		return
	}

	code, err := h.Repo.GetSwiftCodeDetails(swiftCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error retrieving SWIFT code")
		return
	}
	if code == nil {
		writeError(w, http.StatusNotFound, "SWIFT code not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if code.IsHeadquarter {
		branches, err := h.Repo.GetBranchesByHeadquarter(swiftCode)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error retrieving branches")
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
			writeError(w, http.StatusInternalServerError, "Error encoding response")
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
		writeError(w, http.StatusInternalServerError, "Error encoding response")
	}
}

func (h *Handler) GetSwiftCodesByCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	iso2 := vars["countryISO2code"]

	if iso2 == "" {
		writeError(w, http.StatusBadRequest, "Country ISO2 code is required")
		return
	}

	if len(iso2) != 2 {
		writeError(w, http.StatusBadRequest, "Country ISO2 code must be exactly 2 characters")
		return
	}

	codes, countryName, err := h.Repo.GetSwiftCodesByCountry(iso2)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error retrieving SWIFT codes")
		return
	}
	if len(codes) == 0 {
		writeError(w, http.StatusNotFound, "No SWIFT codes found for this country")
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
		writeError(w, http.StatusInternalServerError, "Error encoding response")
		return
	}
}

func (h *Handler) CreateSwiftCode(w http.ResponseWriter, r *http.Request) {
	var req CreateSwiftCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	req.CountryISO2 = strings.ToUpper(req.CountryISO2)
	req.CountryName = strings.ToUpper(req.CountryName)

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
		exists, err := h.Repo.HeadquarterExists(hqCode)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB error")
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
			err = h.Repo.InsertSwiftCodes([]models.SwiftCode{placeholder})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to insert placeholder HQ")
				return
			}
		}

		hqCodePtr := hqCode
		newCode.HeadquarterSWIFTCode = &hqCodePtr
	}

	exists, err := h.Repo.SwiftCodeExists(newCode.SwiftCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB error")
		return
	}

	if exists {
		isPlaceholder, err := h.Repo.IsPlaceholder(newCode.SwiftCode)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB error")
			return
		}
		if isPlaceholder {
			err = h.Repo.UpdatePlaceholderSwiftCode(newCode)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to update placeholder SWIFT code")
				return
			}
		} else {
			writeError(w, http.StatusConflict, "SWIFT code already exists")
			return
		}
	} else {
		if err := h.Repo.InsertSwiftCodes([]models.SwiftCode{newCode}); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to insert SWIFT code")
			return
		}
	}

	writeSuccess(w, "SWIFT code added successfully")
}

func (h *Handler) DeleteSwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]

	if swiftCode == "" {
		writeError(w, http.StatusBadRequest, "SWIFT code is required")
		return
	}

	if len(swiftCode) != 11 {
		writeError(w, http.StatusBadRequest, "SWIFT code must be exactly 11 characters")
		return
	}

	if strings.HasSuffix(swiftCode, "XXX") {
		branches, err := h.Repo.GetBranchesByHeadquarter(swiftCode)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB error")
			return
		}
		if len(branches) > 0 {
			writeError(w, http.StatusConflict, "Cannot delete headquarter with existing branches")
			return
		}
	}

	deleted, err := h.Repo.DeleteSwiftCode(swiftCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error deleting SWIFT code")
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "SWIFT code not found")
		return
	}

	writeSuccess(w, "SWIFT code deleted successfully")
}
