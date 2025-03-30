package routes

import (
	"github.com/gorilla/mux"
	"ethereum-fund-flow/handlers"
)

// SetupRoutes initializes API endpoints
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/beneficiary", handlers.Beneficiary).Methods("GET")
	router.HandleFunc("/payer", handlers.Payer).Methods("GET")

	return router
}