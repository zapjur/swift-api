package main

import (
	"log"
	"net/http"
	"swift-api/internal/database"
	"swift-api/pkg/handlers"
	"swift-api/pkg/repository"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo)

	http.HandleFunc("GET /v1/swift-codes/{swiftCode}", handler.GetSwiftCode)
	http.HandleFunc("GET /v1/swift-codes/country/{countryISO2}", handler.GetSwiftCodesByCountry)
	http.HandleFunc("POST /v1/swift-codes", handler.CreateSwiftCode)
	http.HandleFunc("DELETE /v1/swift-codes/{swiftCode}", handler.DeleteSwiftCode)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
