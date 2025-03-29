package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ethereum-fund-flow/handlers" // Update import path

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)


func TestMain(m *testing.M) {
	// Get absolute path of the project root
	dir, err := os.Getwd()
	if err != nil {
		panic("Error getting working directory")
	}

	// Construct path to .env file
	envPath := dir + "/../.env"

	// Load .env file explicitly
	err = godotenv.Load(envPath)
	if err != nil {
		panic("Error loading .env file in test: " + err.Error())
	}

	// Set API key manually (for safety)
	os.Setenv("ETHERSCAN_API_KEY", os.Getenv("ETHERSCAN_API_KEY"))

	m.Run()
}

func TestBeneficiaryValidAddress(t *testing.T) {
	req, _ := http.NewRequest("GET", "/beneficiary?address=0x742d35Cc6634C0532925a3b844Bc454e4438f44e", nil)
	rec := httptest.NewRecorder()
	handlers.Beneficiary(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}

	if response["message"] != "success" {
		t.Errorf("Expected success message, got %v", response["message"])
	}
}

func TestBeneficiaryInvalidAddress(t *testing.T) {
	req, err := http.NewRequest("GET", "/beneficiary?address=invalidaddress", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.Beneficiary)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBeneficiaryMissingAddress(t *testing.T) {
	req, err := http.NewRequest("GET", "/beneficiary", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.Beneficiary)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}