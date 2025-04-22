package main

import (
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

	repo := repository.NewRepository(db)
	if err = repo.InsertSwiftCodes(hq); err != nil {
		log.Fatal("Error inserting headquarters:", err)
	}

	if err = repo.InsertSwiftCodes(branches); err != nil {
		log.Fatal("Error inserting branches:", err)
	}

	handler := handlers.NewHandler(repo)

	http.HandleFunc("GET /v1/swift-codes/{swiftCode}", handler.GetSwiftCode)
	http.HandleFunc("GET /v1/swift-codes/country/{countryISO2}", handler.GetSwiftCodesByCountry)
	http.HandleFunc("POST /v1/swift-codes", handler.CreateSwiftCode)
	http.HandleFunc("DELETE /v1/swift-codes/{swiftCode}", handler.DeleteSwiftCode)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
