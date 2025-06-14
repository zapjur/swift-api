package handlers_test

import (
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"swift-api/pkg/handlers"
	"swift-api/pkg/repository"
	"testing"
)

func setupTestHandler(t *testing.T) *handlers.Handler {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	h := &handlers.Handler{
		Repo: repository.NewRepository(db),
	}
	_, _ = db.Exec("DELETE FROM swift_codes")
	return h
}

func createHQAndBranch(t *testing.T, h *handlers.Handler) {
	hq := `{"swiftCode":"TSTHQ000XXX","bankName":"HQ","countryISO2":"PL","countryName":"Poland","address":"HQ Addr","isHeadquarter":true}`
	branch := `{"swiftCode":"TSTHQ000001","bankName":"Branch","countryISO2":"PL","countryName":"Poland","address":"Branch Addr","isHeadquarter":false}`

	for _, body := range []string{hq, branch} {
		req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.CreateSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestCreateSwiftCodes(t *testing.T) {
	h := setupTestHandler(t)

	t.Run("Create branch", func(t *testing.T) {
		body := `{"swiftCode":"TSTHQ000001","bankName":"Branch","countryISO2":"PL","countryName":"Poland","address":"Branch Addr","isHeadquarter":false}`
		req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.CreateSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Create HQ", func(t *testing.T) {
		body := `{"swiftCode":"TSTHQ000XXX","bankName":"Test HQ","countryISO2":"PL","countryName":"Poland","address":"HQ St","isHeadquarter":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.CreateSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invalid input - wrong length", func(t *testing.T) {
		body := `{"swiftCode":"SHORT","bankName":"Bad Bank","countryISO2":"PL","countryName":"Poland","address":"Nowhere","isHeadquarter":true}`
		req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.CreateSwiftCode(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Create with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", strings.NewReader(`{invalid-json}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.CreateSwiftCode(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid request payload")
	})
}

func TestGetSwiftCode(t *testing.T) {
	h := setupTestHandler(t)
	createHQAndBranch(t, h)

	t.Run("Get HQ with branches", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/TSTHQ000XXX", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "TSTHQ000XXX"})
		rec := httptest.NewRecorder()
		h.GetSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "TSTHQ000001")
	})

	t.Run("Get branch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/TSTHQ000001", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "TSTHQ000001"})
		rec := httptest.NewRecorder()
		h.GetSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Get non-existent SWIFT code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/INVALIDDXXX", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "INVALIDDXXX"})
		rec := httptest.NewRecorder()
		h.GetSwiftCode(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Get SWIFT code with invalid length", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/SHORT", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "SHORT"})
		rec := httptest.NewRecorder()

		h.GetSwiftCode(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "SWIFT code must be exactly 11 characters")
	})
}

func TestGetSwiftCodesByCountry(t *testing.T) {
	h := setupTestHandler(t)
	createHQAndBranch(t, h)

	t.Run("Get PL codes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/country/PL", nil)
		req = mux.SetURLVars(req, map[string]string{"countryISO2code": "PL"})
		rec := httptest.NewRecorder()
		h.GetSwiftCodesByCountry(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "TSTHQ000XXX")
		assert.Contains(t, rec.Body.String(), "TSTHQ000001")
	})

	t.Run("Get country codes with invalid ISO2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/country/PLS", nil)
		req = mux.SetURLVars(req, map[string]string{"countryISO2code": "PLS"})
		rec := httptest.NewRecorder()

		h.GetSwiftCodesByCountry(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Country ISO2 code must be exactly 2 characters")
	})

	t.Run("Get country codes without swift code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/country/DD", nil)
		req = mux.SetURLVars(req, map[string]string{"countryISO2code": "DD"})
		rec := httptest.NewRecorder()
		h.GetSwiftCodesByCountry(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "No SWIFT codes found for this country")
	})
}

func TestDeleteSwiftCode(t *testing.T) {
	h := setupTestHandler(t)
	createHQAndBranch(t, h)

	t.Run("Delete branch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/swift-codes/TSTHQ000001", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "TSTHQ000001"})
		rec := httptest.NewRecorder()
		h.DeleteSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		reqCheck := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/TSTHQ000001", nil)
		reqCheck = mux.SetURLVars(reqCheck, map[string]string{"swift-code": "TSTHQ000001"})
		recCheck := httptest.NewRecorder()
		h.GetSwiftCode(recCheck, reqCheck)
		assert.Equal(t, http.StatusNotFound, recCheck.Code)
	})

	t.Run("Delete HQ with branch fails", func(t *testing.T) {
		body := `{"swiftCode":"TSTHQ000001","bankName":"Branch Again","countryISO2":"PL","countryName":"Poland","address":"Branch Addr","isHeadquarter":false}`
		req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.CreateSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		reqDel := httptest.NewRequest(http.MethodDelete, "/v1/swift-codes/TSTHQ000XXX", nil)
		reqDel = mux.SetURLVars(reqDel, map[string]string{"swift-code": "TSTHQ000XXX"})
		recDel := httptest.NewRecorder()
		h.DeleteSwiftCode(recDel, reqDel)
		assert.Equal(t, http.StatusConflict, recDel.Code)
	})

	t.Run("Delete HQ after branch deleted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/swift-codes/TSTHQ000001", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "TSTHQ000001"})
		rec := httptest.NewRecorder()
		h.DeleteSwiftCode(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		reqHQ := httptest.NewRequest(http.MethodDelete, "/v1/swift-codes/TSTHQ000XXX", nil)
		reqHQ = mux.SetURLVars(reqHQ, map[string]string{"swift-code": "TSTHQ000XXX"})
		recHQ := httptest.NewRecorder()
		h.DeleteSwiftCode(recHQ, reqHQ)
		assert.Equal(t, http.StatusOK, recHQ.Code)

		reqCheck := httptest.NewRequest(http.MethodGet, "/v1/swift-codes/TSTHQ000XXX", nil)
		reqCheck = mux.SetURLVars(reqCheck, map[string]string{"swift-code": "TSTHQ000XXX"})
		recCheck := httptest.NewRecorder()
		h.GetSwiftCode(recCheck, reqCheck)
		assert.Equal(t, http.StatusNotFound, recCheck.Code)
	})

	t.Run("Delete non-existent SWIFT code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/swift-codes/NONEXISTX01", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": "NONEXISTX01"})
		rec := httptest.NewRecorder()
		h.DeleteSwiftCode(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Delete with empty swift code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/swift-codes/", nil)
		req = mux.SetURLVars(req, map[string]string{"swift-code": ""})
		rec := httptest.NewRecorder()

		h.DeleteSwiftCode(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "SWIFT code is required")
	})

}
