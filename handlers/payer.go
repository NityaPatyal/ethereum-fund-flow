package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"ethereum-fund-flow/utils"
	"ethereum-fund-flow/models"
)

// PayerHandler analyzes sources of funds for a given Ethereum address
func Payer(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")

	if err := utils.ValidateAddress(address); err != nil {
		http.Error(w, "Invalid Ethereum address", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	var wg sync.WaitGroup
	var normalTxs, internalTxs, tokenTxs []models.EtherscanTx
	var err1, err2, err3 error

	wg.Add(3)
	go func() {
		defer wg.Done()
		normalTxs, err1 = utils.FetchTransactions(address, "txlist", apiKey)
	}()
	go func() {
		defer wg.Done()
		internalTxs, err2 = utils.FetchTransactions(address, "txlistinternal", apiKey)
	}()
	go func() {
		defer wg.Done()
		tokenTxs, err3 = utils.FetchTransactions(address, "tokentx", apiKey)
	}()
	wg.Wait()

	if err1 != nil || err2 != nil || err3 != nil {
		http.Error(w, "Error fetching transactions", http.StatusInternalServerError)
		return
	}

	// Analyze incoming transactions (INFLOW)
	payers := utils.AnalyzePayers(normalTxs, internalTxs, tokenTxs, address)

	response := models.APIPayerResponse{Message: "success", Data: payers}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}