package main

import (
	deliveryHTTP "calendar_app/internal/delivery/http"
	"log"
	"net/http"
)

func main() {
	router := deliveryHTTP.NewRouter()

	log.Println("Starting calendar server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
