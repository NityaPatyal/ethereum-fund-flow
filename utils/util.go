package utils

import (
	"encoding/json"
	"errors"
	"ethereum-fund-flow/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
        return []models.EtherscanTx{}, nil // Return empty slice instead of an error
    }

    // Convert interface{} to []models.EtherscanTx
    resultJSON, _ := json.Marshal(etherscanResp.Result)
    var transactions []models.EtherscanTx
    if err := json.Unmarshal(resultJSON, &transactions); err != nil {
        return nil, fmt.Errorf("failed to parse transactions: %v", err)
    }

    return transactions, nil
}


// AnalyzeTransactions processes transactions to find beneficiaries
func AnalyzeTransactions(normalTxs, internalTxs, tokenTxs []models.EtherscanTx) []models.Beneficiary {
	beneficiaryMap := make(map[string]*models.Beneficiary)

	// Function to process each transaction
	processTx := func(tx models.EtherscanTx, isToken bool, isInternal bool) {
		to := tx.To
		amount, _ := strconv.ParseFloat(tx.Value, 64)
		if isToken {
			amount, _ = strconv.ParseFloat(tx.Value, 64)
		}

		// If internal transaction, check for smart contract involvement
		if isInternal && IsContractAddress(to) {
			finalBeneficiary := TraceFinalBeneficiary(to, internalTxs)
			if finalBeneficiary != "" {
				to = finalBeneficiary // Set the final traced recipient
			}
		}

		// Ensure valid Ethereum address
		if common.IsHexAddress(to) {
			if _, exists := beneficiaryMap[to]; !exists {
				beneficiaryMap[to] = &models.Beneficiary{
					Address:      to,
					Amount:       0,
					Transactions: []models.TxInfo{},
				}
			}

			// Convert timestamp and append transaction info
			timestamp, _ := strconv.ParseInt(tx.TimeStamp, 10, 64)
			beneficiaryMap[to].Amount += amount
			beneficiaryMap[to].Transactions = append(beneficiaryMap[to].Transactions, models.TxInfo{
				TransactionID: tx.Hash,
				TxAmount:      amount,
				DateTime:      time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
			})
		}
	}

	// Process all three transaction types
	for _, tx := range normalTxs {
		processTx(tx, false, false) // Normal transaction (ETH transfer)
	}
	for _, tx := range internalTxs {
		processTx(tx, false, true) // Internal transaction (ETH transfer with contract check)
	}
	for _, tx := range tokenTxs {
		processTx(tx, true, false) // Token transfer
	}

	// Convert map to slice
	var beneficiaries []models.Beneficiary
	for _, b := range beneficiaryMap {
		beneficiaries = append(beneficiaries, *b)
	}

	return beneficiaries
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
        log.Printf("Error parsing timestamp for Tx %s: %v", tx.Hash, err)
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

// TraceFinalBeneficiary finds the final non-contract recipient from a series of internal transactions
func TraceFinalBeneficiary(contractAddr string, internalTxs []models.EtherscanTx) string {
	for _, tx := range internalTxs {
		if tx.From == contractAddr {
			if !IsContractAddress(tx.To) {
				return tx.To // Found final beneficiary
			}
			return TraceFinalBeneficiary(tx.To, internalTxs) // Recursively trace
		}
	}
	return contractAddr // If no further transfers, return original address
}

// IsContractAddress checks if the given Ethereum address is a smart contract
func IsContractAddress(address string) bool {
	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	if apiKey == "" {
		fmt.Println("Etherscan API Key not set!")
		return false
	}

	url := fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s", address, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching contract info:", err)
		return false
	}
	defer resp.Body.Close()

	var etherscanResp models.EtherscanContractAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&etherscanResp); err != nil {
		fmt.Println("Error decoding response:", err)
		return false
	}

	// If the result is "Contract source code not verified", it's likely not a contract
	return etherscanResp.Result != "Contract source code not verified" && etherscanResp.Result != ""
}