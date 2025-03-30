# Ethereum Fund Flow Analysis

## Overview
This project provides a Go-based API to analyze the flow of funds for a given Ethereum address. It fetches transactions using the Etherscan API and determines the beneficiary addresses.

## Features
- Fetches Normal, Internal, and Token Transactions (ERC-20, ERC-721, ERC-1155).
- Identifies ultimate beneficiaries in transaction chains.
- Optimized API calls with concurrency.
- Structured logging and error handling.

## Prerequisites
- Go 1.18+
- A valid [Etherscan API Key](https://etherscan.io/myapikey)
- Git

## Project Structure
```
.
├── README.md
├── go.mod
├── go.sum
├── handlers
│   └── beneficiary.go
│   └── payer.go
├── main.go
├── models
│   └── types.go
├── routes
│   └── routes.go
├── tests
│   └── beneficiary_test.go
│   └── payer_test.go
└── utils
    └── util.go
```

## Setup Instructions

1. **Clone the Repository**
   ```sh
   git clone <repository-url>
   cd ethereum-fund-flow
   ```

## Installation & Setup

### 2. Install Dependencies
```sh
go mod tidy
```

### 3. Set Up Environment Variables
Create a `.env` file in the root directory and add your Etherscan API key:
```sh
ETHERSCAN_API_KEY=your_etherscan_api_key_here
```

### 4. Run the API
```sh
go run main.go
```
The server will start on [http://localhost:8080](http://localhost:8080)

---

## API Endpoints

### 1. Get Beneficiary Data

Hit the API endpoint using curl command in separate terminal

**Request:**
```sh
curl -X GET "http://localhost:8080/beneficiary?address=<address>" -H "Content-Type: application/json" > <file-path to save response>
```

Example:
```sh
curl -X GET "http://localhost:8080/beneficiary?address=0x1218E12D77A8D1ad56Ec2f6d3d09A428cb7FDA7c" -H "Content-Type: application/json" > output.json
```

**Response:**
```json
{
  "message": "success",
  "data": [
    {
      "beneficiary_address": "0x...",
      "amount": 1.5,
      "transactions": [
        {
          "tx_amount": 1.5,
          "date_time": "YYYY-MM-DD HH:MM:SS",
          "transaction_id": "0x..."
        }
      ]
    }
  ]
}
```


### 2. Get Payer Data

Hit the API endpoint using curl command in separate terminal

**Request:**
```sh
curl -X GET "http://localhost:8080/payer?address=<address>" -H "Content-Type: application/json" > <file-path to save response>
```

Example:
```sh
curl -X GET "http://localhost:8080/payer?address=0x1218E12D77A8D1ad56Ec2f6d3d09A428cb7FDA7c" -H "Content-Type: application/json" > output.json
```

**Response:**
```json
{
  "message": "success",
  "data": [
    {
      "payer_address": "0x...",
      "amount": 1.5,
      "transactions": [
        {
          "tx_amount": 1.5,
          "date_time": "YYYY-MM-DD HH:MM:SS",
          "transaction_id": "0x..."
        }
      ]
    }
  ]
}
```

---


## Running Tests

### 1. Run Unit Tests
```sh
go test -v ./tests/
```

### 2. Common Test Issues & Fixes
- **Error: Invalid API Key** → Ensure `.env` file is properly loaded and contains a valid key.
- **Test Fails for Invalid Address** → Verify API response format.

---

## Contributing

1. Fork the repository
2. Create a new branch
   ```sh
   git checkout -b feature-branch
   ```
3. Commit your changes
   ```sh
   git commit -m "Added new feature"
   ```
4. Push the branch
   ```sh
   git push origin feature-branch
   ```
5. Create a Pull Request

---

