package models  

// Beneficiary represents a single beneficiary structure
type Beneficiary struct {
    Address  string `json:"beneficiary_address"`
	Amount       float64  `json:"amount"`
	Transactions []TxInfo `json:"transactions"`
}

// TxInfo represents transaction details
type TxInfo struct {
	Amount        float64 `json:"tx_amount"`
	DateTime      string  `json:"date_time"`
	TransactionID string  `json:"transaction_id"`
}

// Response structure
type APIResponse struct {
	Message string       `json:"message"`
	Data    []Beneficiary `json:"data"`
}

// Etherscan response structures
type EtherscanTx struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
	Hash   string `json:"hash"`
	Time   string `json:"timeStamp"`
}

// Etherscan response structures
type EtherscanResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []EtherscanTx `json:"result"`
}

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

type APIPayerResponse struct {
	Message string       `json:"message"`
	Data    []Payer `json:"data"`
}