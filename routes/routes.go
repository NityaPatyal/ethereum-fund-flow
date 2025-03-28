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

	return router
}