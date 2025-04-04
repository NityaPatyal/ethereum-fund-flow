package models  

// Beneficiary represents a single beneficiary structure
type Beneficiary struct {
    Address      string  `json:"beneficiary_address"`
	Amount       float64 `json:"amount"`
	Transactions []TxInfo `json:"transactions"`
}

// TxInfo represents transaction details
type TxInfo struct {
	TransactionID string  `json:"transaction_id"`
	TxAmount      float64 `json:"tx_amount"`
	DateTime      string  `json:"date_time"`
}

// API Response structure for Beneficiary
type APIResponse struct {
	Message string        `json:"message"`
	Data    []Beneficiary `json:"data"`
}

// Etherscan transaction structure
// EtherscanTx represents a transaction from Etherscan API response
type EtherscanTx struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
	Hash   string `json:"hash"`
	TimeStamp string `json:"timeStamp"` 
}

// Etherscan response structure
type EtherscanResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []EtherscanTx `json:"result"`
}

// EtherscanContractAPIResponse represents the response from Etherscan API for checking Contract Address
type EtherscanContractAPIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// Payer-related structs
type Payer struct {
	PayerAddress string             `json:"payer_address"`
	Amount       float64            `json:"amount"`
	Transactions []PayerTransaction `json:"transactions"`
}

type PayerTransaction struct {
	TxAmount      float64 `json:"tx_amount"`
	DateTime      string  `json:"date_time"`
	TransactionID string  `json:"transaction_id"`
}

// Payer API response structure
type APIPayerResponse struct {
	Message string  `json:"message"`
	Data    []Payer `json:"data"`
}