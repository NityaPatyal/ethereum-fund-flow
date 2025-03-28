package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"ethereum-fund-flow/utils"
	"ethereum-fund-flow/models"
)

// BeneficiaryHandler handles API requests
func Beneficiary(w http.ResponseWriter, r *http.Request){
    address :=r.URL.Query().Get("address")
    if address == "" {
        http.Error(w, "Missing address parameter", http.StatusBadRequest)
        return
    }

    apiKey := os.Getenv("ETHERSCAN_API_KEY")
	var wg sync.WaitGroup
	var normalTxs, internalTxs, tokenTxs []models.EtherscanTx
	var err1, err2, err3 error
	var status1, status2, status3 int
	
	wg.Add(3)
	go func() {
		defer wg.Done()
		normalTxs, status1, err1 = utils.FetchTransactions(address, "txlist", apiKey)
	}()
	go func() {
		defer wg.Done()
		internalTxs, status2, err2 = utils.FetchTransactions(address, "txlistinternal", apiKey)
	}()
	go func() {
		defer wg.Done()
		tokenTxs, status3, err3 = utils.FetchTransactions(address, "tokentx", apiKey)
	}()
	wg.Wait()
	
	if err1 != nil || err2 != nil || err3 != nil {
        status := http.StatusInternalServerError
        if status1 == http.StatusBadRequest || status2 == http.StatusBadRequest || status3 == http.StatusBadRequest {
            status = http.StatusBadRequest
        }
        http.Error(w, "Error fetching transactions", status)
        return
    }
	
	beneficiaries := utils.AnalyzeTransactions(normalTxs, internalTxs, tokenTxs) 
	response := models.APIResponse{Message: "success", Data: beneficiaries}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}