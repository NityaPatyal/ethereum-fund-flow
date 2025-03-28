package utils

import (
	"encoding/json"
	"ethereum-fund-flow/models"
	"fmt"
	"io"
	"net/http"
)

const etherscanAPI = "https://api.etherscan.io/api"

// FetchTransactions fetches transactions from Etherscan
func FetchTransactions(address, action, apiKey string) ([]models.EtherscanTx, int, error) {
    url := fmt.Sprintf("%s?module=account&action=%s&address=%s&apikey=%s", etherscanAPI, action, address, apiKey)
    resp, err := http.Get(url)
    if err != nil {
        return nil, http.StatusInternalServerError, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, http.StatusInternalServerError, err
    }

	var etherscanResp struct {
        Status  string      `json:"status"`
        Message string      `json:"message"`
        Result  interface{} `json:"result"`
    }

    if err := json.Unmarshal(body, &etherscanResp); err != nil {
        return nil, http.StatusInternalServerError, err
    }

    if etherscanResp.Status != "1" {
        if msg, ok := etherscanResp.Result.(string); ok && msg == "Invalid Address format" {
            return nil, http.StatusBadRequest, fmt.Errorf("invalid Ethereum address")
        }
        return nil, http.StatusInternalServerError, fmt.Errorf("etherscan API error: %s", etherscanResp.Message)
    }

	// ✅ Safely convert result to transactions slice
    transactions, ok := etherscanResp.Result.([]models.EtherscanTx)
    if !ok {
        return nil, http.StatusInternalServerError, fmt.Errorf("unexpected response format from Etherscan")
    }

    return transactions, http.StatusOK, nil }

// AnalyzeTransactions determines beneficiaries recursively
func AnalyzeTransactions(normal, internal, token []models.EtherscanTx) []models.Beneficiary {
	// Parse transactions and build a transaction graph
	txGraph := make(map[string][]models.TxInfo)
	// Parse normal, internal, and token transactions, add to txGraph
	processTransactions(normal, txGraph, "Normal")
	processTransactions(internal, txGraph, "Internal")
	processTransactions(token, txGraph, "Token")


	// Find last recipients (ultimate beneficiaries)
	beneficiaries := findUltimateBeneficiaries(txGraph)
	return beneficiaries
}

// findUltimateBeneficiaries finds the last recipients in a transaction chain
func findUltimateBeneficiaries(txGraph map[string][]models.TxInfo) []models.Beneficiary {
	visited := make(map[string]bool)
	var results []models.Beneficiary

	for _, transactions := range txGraph {
		for _, tx := range transactions {
			if !visited[tx.TransactionID] {
				visited[tx.TransactionID] = true
				results = append(results, models.Beneficiary{
					Address: tx.TransactionID,
					Amount:  tx.Amount,
					Transactions: []models.TxInfo{tx},
				})
			}
		}
	}
	return results
}

// processTransactions fills txGraph from transaction data
func processTransactions(transactions []models.EtherscanTx, txGraph map[string][]models.TxInfo, txType string) {
    for _, tx := range transactions {
        if tx.To == "" {
            continue // Ignore transactions with no recipient
        }

        txInfo := models.TxInfo{
            Amount:        parseValue(tx.Value),
            DateTime:      tx.Time,
            TransactionID: tx.Hash,
        }

        // Differentiate between transaction types
        if txType == "Token" {
            txGraph[tx.From] = append(txGraph[tx.From], txInfo) // Track token movements
        } else {
            txGraph[tx.To] = append(txGraph[tx.To], txInfo) // Track fund flow to recipients
        }
    }
}

// parseValue converts string value to float64
func parseValue(value string) float64 {
	var amount float64
	fmt.Sscanf(value, "%f", &amount)
	return amount / 1e18 // Convert from Wei to ETH
}
