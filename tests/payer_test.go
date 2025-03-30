package tests

import (
	"ethereum-fund-flow/models"
	"ethereum-fund-flow/utils"
	"testing"
	"time"
)

func TestAnalyzePayers(t *testing.T) {
	// ✅ Dummy transaction data
	transactions := []models.EtherscanTx{
		{From: "0xSender1", To: "0xTarget", Value: "1.5", Hash: "0xTx1", TimeStamp: "1711711711"},
		{From: "0xSender2", To: "0xTarget", Value: "2.0", Hash: "0xTx2", TimeStamp: "1711711712"},
		{From: "0xSender1", To: "0xTarget", Value: "0.5", Hash: "0xTx3", TimeStamp: "1711711713"},
	}

	// ✅ Call function
	targetAddress := "0xTarget"
	payers := utils.AnalyzePayers(transactions, nil, nil, targetAddress)

	// ✅ Expecting 2 payers
	if len(payers) != 2 {
		t.Errorf("Expected 2 payers, got %d", len(payers))
	}

	// ✅ Check first payer (0xSender1)
	payer1 := payers[0]
	if payer1.PayerAddress != "0xSender1" {
		t.Errorf("Expected payer address 0xSender1, got %s", payer1.PayerAddress)
	}
	if payer1.Amount != 2.0 {
		t.Errorf("Expected amount 2.0, got %f", payer1.Amount)
	}
	if len(payer1.Transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(payer1.Transactions))
	}

	// ✅ Check timestamp format
	expectedTime := time.Unix(1711711711, 0).Format("2006-01-02 15:04:05")
	if payer1.Transactions[0].DateTime != expectedTime {
		t.Errorf("Expected date %s, got %s", expectedTime, payer1.Transactions[0].DateTime)
	}
}