package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
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

	if code.IsHeadquarter {
		branches, err := h.repo.GetBranchesByHeadquarter(swiftCode)
		if err != nil {
			http.Error(w, "Error retrieving branches", http.StatusInternalServerError)
			return
		}
		code.Branches = branches
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(code); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetSwiftCodesByCountry(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) CreateSwiftCode(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) DeleteSwiftCode(w http.ResponseWriter, r *http.Request) {

}
