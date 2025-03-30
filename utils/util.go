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
func AnalyzeTransactions(normalTxs, internalTxs, tokenTxs []models.EtherscanTx) []models.Beneficiary {
	txGraph := make(map[string][]models.TxInfo) // Sender → List of Transactions
	seenRecipients := make(map[string]bool)     // Track all recipients
	allSenders := make(map[string]bool)         // Track senders

	// Process Transactions
	processTransactions(normalTxs, txGraph, seenRecipients, allSenders)
	processTransactions(internalTxs, txGraph, seenRecipients, allSenders)
	processTransactions(tokenTxs, txGraph, seenRecipients, allSenders)

	// Find Beneficiaries (Addresses that receive funds but don't send)
	var beneficiaries []models.Beneficiary
	for recipient := range seenRecipients {
		if !allSenders[recipient] { // If recipient is NOT a sender, it's a beneficiary
			beneficiaries = append(beneficiaries, models.Beneficiary{
				Address:      recipient,
				Amount:       calculateTotalAmount(txGraph[recipient]),
				Transactions: txGraph[recipient],
			})
		}
	}

	return beneficiaries
}

// processTransactions processes a list of transactions and populates the transaction graph
func processTransactions(txs []models.EtherscanTx, txGraph map[string][]models.TxInfo, seenRecipients, allSenders map[string]bool) {
	for _, tx := range txs {
		amount, _ := strconv.ParseFloat(tx.Value, 64)
		if amount <= 0 {
			continue // Ignore zero-value transactions
		}

		txInfo := models.TxInfo{
			TransactionID: tx.Hash,
			TxAmount:      amount,
			DateTime:      parseTimestamp(tx.TimeStamp),
		}

		// Add transaction to sender's record
		txGraph[tx.To] = append(txGraph[tx.To], txInfo)

		// Mark the recipient as seen
		seenRecipients[tx.To] = true

		// Track senders separately
		allSenders[tx.From] = true
	}
}

// calculateTotalAmount sums up transaction amounts for a given address
func calculateTotalAmount(transactions []models.TxInfo) float64 {
	var total float64
	for _, tx := range transactions {
		total += tx.TxAmount
	}
	return total
}

func AnalyzePayers(normalTxs, internalTxs, tokenTxs []models.EtherscanTx, targetAddress string) []models.Payer {
    payerMap := make(map[string]*models.Payer)

    // Step 1: Direct Transactions (Where to == targetAddress)
    directSenders := make(map[string]bool)
	processTransaction := func(tx models.EtherscanTx) {

		if strings.EqualFold(tx.To, targetAddress) {
	
			directSenders[tx.From] = true
			addPayer(tx, payerMap)
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

    // Step 2: Backtrack - Find where the direct senders got their funds
    for payer := range directSenders {
        trackFunds(payer, normalTxs, internalTxs, tokenTxs, payerMap)
    }

    var payers []models.Payer
    for _, payer := range payerMap {
        payers = append(payers, *payer)
    }

    return payers
}

// Helper function to add payers to the map
func addPayer(tx models.EtherscanTx, payerMap map[string]*models.Payer) {
    amount, err := strconv.ParseFloat(tx.Value, 64)
    if err != nil {
        log.Printf("❌ Error parsing amount for Tx %s: %v", tx.Hash, err)
        return
    }

    if _, exists := payerMap[tx.From]; !exists {
        payerMap[tx.From] = &models.Payer{
            PayerAddress: tx.From,
            Amount:       0.0,
            Transactions: []models.PayerTransaction{},
        }
    }

    timestampInt, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
    if err != nil {
        log.Printf("❌ Error parsing timestamp for Tx %s: %v", tx.Hash, err)
        timestampInt = 0
    }

    payerMap[tx.From].Amount += amount
    payerMap[tx.From].Transactions = append(payerMap[tx.From].Transactions, models.PayerTransaction{
        TxAmount:      amount,
        DateTime:      time.Unix(timestampInt, 0).Format("2006-01-02 15:04:05"),
        TransactionID: tx.Hash,
    })
}

// Backtrack to find where the payer got their funds
func trackFunds(payer string, normalTxs, internalTxs, tokenTxs []models.EtherscanTx, payerMap map[string]*models.Payer) {
    for _, tx := range normalTxs {
        if tx.To == payer {
            addPayer(tx, payerMap)
        }
    }
    for _, tx := range internalTxs {
        if tx.To == payer {
            addPayer(tx, payerMap)
        }
    }
    for _, tx := range tokenTxs {
        if tx.To == payer {
            addPayer(tx, payerMap)
        }
    }
}

func parseTimestamp(timestamp string) string {
    ts, err := strconv.ParseInt(timestamp, 10, 64)
    if err != nil {
        return "" // Handle error gracefully
    }
    return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
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
