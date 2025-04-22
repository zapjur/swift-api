package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"swift-api/internal/database"
	"swift-api/pkg/handlers"
	"swift-api/pkg/parser"
	"swift-api/pkg/repository"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	filePath := os.Getenv("SWIFT_CODES_FILE_PATH")
	if filePath == "" {
		filePath = "./assets/swift_codes.csv"
	}

	hq, branches, err := parser.ParseCSV(filePath)
	if err != nil {
		log.Fatal(err)
	}
	hq = parser.FillMissingHeadquarters(hq, branches)

	repo := repository.NewRepository(db)
	if err = repo.InsertSwiftCodes(hq); err != nil {
		log.Fatal("Error inserting headquarters:", err)
	}

	if err = repo.InsertSwiftCodes(branches); err != nil {
		log.Fatal("Error inserting branches:", err)
	}

	handler := handlers.NewHandler(repo)

	r := mux.NewRouter()

	r.HandleFunc("/v1/swift-codes/{swift-code}", handler.GetSwiftCode).Methods("GET")
	r.HandleFunc("/v1/swift-codes/country/{countryISO2code}", handler.GetSwiftCodesByCountry).Methods("GET")
	r.HandleFunc("/v1/swift-codes", handler.CreateSwiftCode).Methods("POST")
	r.HandleFunc("/v1/swift-codes/{swift-code}", handler.DeleteSwiftCode).Methods("DELETE")

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
