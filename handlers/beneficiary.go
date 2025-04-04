package handlers

import (
	"encoding/json"
	"ethereum-fund-flow/models"
	"ethereum-fund-flow/utils"
	"net/http"
	"os"
	"sync"
)

// BeneficiaryHandler handles API requests
func Beneficiary(w http.ResponseWriter, r *http.Request){
    address :=r.URL.Query().Get("address")

	if err := utils.ValidateAddress(address) ; err != nil {
        http.Error(w, "Invalid Ethereum address", http.StatusBadRequest) // 🛑 400 error
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
		status := http.StatusInternalServerError
		http.Error(w, "Error fetching transactions", status)
		return
	}
	
	beneficiaries := utils.AnalyzeTransactions(normalTxs, internalTxs, tokenTxs) 
	response := models.APIResponse{Message: "success", Data: beneficiaries}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}