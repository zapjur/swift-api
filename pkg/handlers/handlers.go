package handlers

import (
	"encoding/json"
	"net/http"
	"swift-api/pkg/repository"
)

type Handler struct {
	repo repository.Repository
}

func NewHandler(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetSwiftCode(w http.ResponseWriter, r *http.Request) {
	swiftCode := r.PathValue("swiftCode")

	json.NewEncoder(w).Encode(map[string]string{"swiftCode": swiftCode})
}

func (h *Handler) GetSwiftCodesByCountry(w http.ResponseWriter, r *http.Request) {
	countryISO2 := r.PathValue("countryISO2")

	json.NewEncoder(w).Encode(map[string]string{"countryISO2": countryISO2})
}

func (h *Handler) CreateSwiftCode(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Swift code created"})
}

func (h *Handler) DeleteSwiftCode(w http.ResponseWriter, r *http.Request) {
	swiftCode := r.PathValue("swiftCode")

	json.NewEncoder(w).Encode(map[string]string{"message": "Swift code " + swiftCode + " deleted"})
}
