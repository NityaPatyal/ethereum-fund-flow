package utils

import (
	"encoding/json"
	"errors"
	"ethereum-fund-flow/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const etherscanAPI = "https://api.etherscan.io/api"

// FetchTransactions fetches transactions from Etherscan
func FetchTransactions(address, action, apiKey string) ([]models.EtherscanTx, error) {
    url := fmt.Sprintf("%s?module=account&action=%s&address=%s&apikey=%s", etherscanAPI, action, address, apiKey)
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %v", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }

    var etherscanResp struct {
        Status  string      `json:"status"`
        Message string      `json:"message"`
        Result  interface{} `json:"result"`
    }

    if err := json.Unmarshal(body, &etherscanResp); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %v", err)
    }

    if etherscanResp.Status != "1" {
        errMsg := fmt.Sprintf("Etherscan API error: %s - %v", etherscanResp.Message, etherscanResp.Result)
        return nil, fmt.Errorf(errMsg)
    }

    // Convert interface{} to []models.EtherscanTx
    resultJSON, _ := json.Marshal(etherscanResp.Result)
    var transactions []models.EtherscanTx
    if err := json.Unmarshal(resultJSON, &transactions); err != nil {
        return nil, fmt.Errorf("failed to parse transactions: %v", err)
    }

    return transactions, nil
}

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

func AnalyzePayers(normalTxs, internalTxs, tokenTxs []models.EtherscanTx, targetAddress string) []models.Payer {
	payerMap := make(map[string]*models.Payer)

	processTransaction := func(tx models.EtherscanTx) {
		if tx.To == targetAddress {
			amount, err := strconv.ParseFloat(tx.Value, 64)
			if err != nil {
				log.Printf("Error parsing amount: %v", err)
				return
			}

			if _, exists := payerMap[tx.From]; !exists {
				payerMap[tx.From] = &models.Payer{
					PayerAddress: tx.From,
					Amount:       0.0,
					Transactions: []models.PayerTransaction{},
				}
			}

			// ✅ Fix: Convert timestamp to int64
			timestampInt, err := strconv.ParseInt(tx.Time, 10, 64)
			if err != nil {
				log.Printf("Error parsing timestamp: %v", err)
				timestampInt = 0 // Default value
			}

			// ✅ Fix: Ensure Amount is float64
			payerMap[tx.From].Amount += amount

			payerMap[tx.From].Transactions = append(payerMap[tx.From].Transactions, models.PayerTransaction{
				TxAmount:      amount,
				DateTime:      time.Unix(timestampInt, 0).Format("2006-01-02 15:04:05"),
				TransactionID: tx.Hash,
			})
		}
	}

	for _, tx := range normalTxs {
		processTransaction(tx)
	}
	for _, tx := range internalTxs {
		processTransaction(tx)
	}
	for _, tx := range tokenTxs {
		processTransaction(tx)
	}

	var payers []models.Payer
	for _, payer := range payerMap {
		payers = append(payers, *payer)
	}

	return payers
}

// parseValue converts string value to float64
func parseValue(value string) float64 {
	var amount float64
	fmt.Sscanf(value, "%f", &amount)
	return amount / 1e18 // Convert from Wei to ETH
}


// Ethereum address validate karne ka function
func isValidEthereumAddress(address string) bool {
	// Address converted to lowercase 
	address = strings.ToLower(address)

	// Ethereum address regex pattern
	re := regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

	// Check is matching address
	return re.MatchString(address)
}


func ValidateAddress(address string) error {
    if !isValidEthereumAddress(address) {
        return errors.New("invalid Ethereum address")
    }
    return nil
}