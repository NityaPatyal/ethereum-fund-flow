package main

import (
	"ethereum-fund-flow/routes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Get API key from environment variables
	etherscanAPIKey := os.Getenv("ETHERSCAN_API_KEY")
	if etherscanAPIKey == "" {
		log.Fatal("Missing ETHERSCAN_API_KEY in .env file")
	}

	// Create a new router
	router := routes.SetupRoutes()
	// Start the server
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
	
}