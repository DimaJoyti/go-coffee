# Web3 Wallet Backend API Documentation

This document provides detailed information about the Web3 Wallet Backend API endpoints, request/response formats, and usage examples.

## Table of Contents

1. [Authentication](#authentication)
2. [Wallet API](#wallet-api)
3. [Transaction API](#transaction-api)
4. [Smart Contract API](#smart-contract-api)
5. [Security API](#security-api)
6. [Error Handling](#error-handling)

## Authentication

All API requests require authentication using JWT tokens.

### Obtain a Token

```
POST /api/v1/auth/login
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1620000000
}
```

### Refresh a Token

```
POST /api/v1/auth/refresh
```

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1620000000
}
```

## Wallet API

### Create a Wallet

```
POST /api/v1/wallets
```

**Request Body:**

```json
{
  "name": "My Ethereum Wallet",
  "chain": "ethereum",
  "type": "hd"
}
```

**Response:**

```json
{
  "wallet": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Ethereum Wallet",
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "chain": "ethereum",
    "type": "hd",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  },
  "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
  "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "derivation_path": "m/44'/60'/0'/0/0"
}
```

### Get a Wallet

```
GET /api/v1/wallets/{id}
```

**Response:**

```json
{
  "wallet": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Ethereum Wallet",
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "chain": "ethereum",
    "type": "hd",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### List Wallets

```
GET /api/v1/wallets
```

**Query Parameters:**

- `chain` (optional): Filter by blockchain (ethereum, bsc, polygon)
- `type` (optional): Filter by wallet type (hd, imported, multisig)
- `limit` (optional): Number of results per page (default: 10)
- `offset` (optional): Pagination offset (default: 0)

**Response:**

```json
{
  "wallets": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "My Ethereum Wallet",
      "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "chain": "ethereum",
      "type": "hd",
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

### Get Wallet Balance

```
GET /api/v1/wallets/{id}/balance
```

**Query Parameters:**

- `token_address` (optional): ERC-20 token address (if not provided, returns native token balance)

**Response:**

```json
{
  "balance": "1000000000000000000",
  "symbol": "ETH",
  "decimals": 18,
  "token_address": ""
}
```

### Import a Wallet

```
POST /api/v1/wallets/import
```

**Request Body:**

```json
{
  "name": "Imported Wallet",
  "chain": "ethereum",
  "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}
```

**Response:**

```json
{
  "wallet": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Imported Wallet",
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "chain": "ethereum",
    "type": "imported",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### Export a Wallet

```
POST /api/v1/wallets/{id}/export
```

**Request Body:**

```json
{
  "passphrase": "your-secure-passphrase"
}
```

**Response:**

```json
{
  "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "keystore": "{\"version\":3,\"id\":\"7e59dc02-8d42-409d-b29a-a8a0f862cc81\",\"address\":\"742d35cc6634c0532925a3b844bc454e4438f44e\",\"crypto\":{...}}"
}
```

### Delete a Wallet

```
DELETE /api/v1/wallets/{id}
```

**Response:**

```json
{
  "success": true
}
```

## Transaction API

### Create a Transaction

```
POST /api/v1/transactions
```

**Request Body:**

```json
{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
  "value": "1000000000000000000",
  "gas": 21000,
  "gas_price": "20000000000",
  "data": "0x",
  "passphrase": "your-secure-passphrase"
}
```

**Response:**

```json
{
  "transaction": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
    "hash": "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
    "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "value": "1000000000000000000",
    "gas": 21000,
    "gas_price": "20000000000",
    "nonce": 0,
    "data": "0x",
    "chain": "ethereum",
    "status": "pending",
    "block_number": 0,
    "block_hash": "",
    "confirmations": 0,
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### Get a Transaction

```
GET /api/v1/transactions/{id}
```

**Response:**

```json
{
  "transaction": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
    "hash": "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
    "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "value": "1000000000000000000",
    "gas": 21000,
    "gas_price": "20000000000",
    "nonce": 0,
    "data": "0x",
    "chain": "ethereum",
    "status": "confirmed",
    "block_number": 12345678,
    "block_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    "confirmations": 10,
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### List Transactions

```
GET /api/v1/transactions
```

**Query Parameters:**

- `wallet_id` (optional): Filter by wallet ID
- `status` (optional): Filter by status (pending, confirmed, failed)
- `chain` (optional): Filter by blockchain (ethereum, bsc, polygon)
- `limit` (optional): Number of results per page (default: 10)
- `offset` (optional): Pagination offset (default: 0)

**Response:**

```json
{
  "transactions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
      "hash": "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
      "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "value": "1000000000000000000",
      "gas": 21000,
      "gas_price": "20000000000",
      "nonce": 0,
      "data": "0x",
      "chain": "ethereum",
      "status": "confirmed",
      "block_number": 12345678,
      "block_hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "confirmations": 10,
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

### Estimate Gas

```
POST /api/v1/transactions/estimate-gas
```

**Request Body:**

```json
{
  "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
  "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
  "value": "1000000000000000000",
  "data": "0x",
  "chain": "ethereum"
}
```

**Response:**

```json
{
  "gas": 21000
}
```

### Get Gas Price

```
GET /api/v1/transactions/gas-price
```

**Query Parameters:**

- `chain` (required): Blockchain (ethereum, bsc, polygon)

**Response:**

```json
{
  "gas_price": "20000000000",
  "slow": "16000000000",
  "average": "20000000000",
  "fast": "24000000000"
}
```

## Smart Contract API

### Deploy a Contract

```
POST /api/v1/contracts
```

**Request Body:**

```json
{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Token",
  "chain": "ethereum",
  "type": "erc20",
  "abi": "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]",
  "bytecode": "0x60806040523480156100...",
  "arguments": ["My Token", "MTK"],
  "gas": 4000000,
  "gas_price": "20000000000",
  "passphrase": "your-secure-passphrase"
}
```

**Response:**

```json
{
  "contract": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Token",
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "chain": "ethereum",
    "abi": "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]",
    "bytecode": "0x60806040523480156100...",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  },
  "transaction": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "hash": "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
    "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "to": "",
    "value": "0",
    "gas": 4000000,
    "gas_price": "20000000000",
    "status": "pending"
  }
}
```

### Import a Contract

```
POST /api/v1/contracts/import
```

**Request Body:**

```json
{
  "name": "Imported Token",
  "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
  "chain": "ethereum",
  "type": "erc20",
  "abi": "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"
}
```

**Response:**

```json
{
  "contract": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Imported Token",
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "chain": "ethereum",
    "abi": "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]",
    "bytecode": "",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### Get a Contract

```
GET /api/v1/contracts/{id}
```

**Response:**

```json
{
  "contract": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Token",
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "chain": "ethereum",
    "abi": "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]",
    "bytecode": "0x60806040523480156100...",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### List Contracts

```
GET /api/v1/contracts
```

**Query Parameters:**

- `chain` (optional): Filter by blockchain (ethereum, bsc, polygon)
- `type` (optional): Filter by contract type (erc20, erc721, erc1155, custom)
- `limit` (optional): Number of results per page (default: 10)
- `offset` (optional): Pagination offset (default: 0)

**Response:**

```json
{
  "contracts": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "My Token",
      "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "chain": "ethereum",
      "abi": "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]",
      "bytecode": "0x60806040523480156100...",
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

### Call a Contract Method (Read-Only)

```
POST /api/v1/contracts/{id}/call
```

**Request Body:**

```json
{
  "method": "balanceOf",
  "arguments": ["0x742d35Cc6634C0532925a3b844Bc454e4438f44e"],
  "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
}
```

**Response:**

```json
{
  "result": "1000000000000000000"
}
```

### Send a Contract Transaction (State-Changing)

```
POST /api/v1/contracts/{id}/send
```

**Request Body:**

```json
{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "transfer",
  "arguments": ["0x742d35Cc6634C0532925a3b844Bc454e4438f44e", "1000000000000000000"],
  "value": "0",
  "gas": 100000,
  "gas_price": "20000000000",
  "passphrase": "your-secure-passphrase"
}
```

**Response:**

```json
{
  "transaction": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
    "hash": "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
    "from": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "value": "0",
    "gas": 100000,
    "gas_price": "20000000000",
    "nonce": 1,
    "data": "0xa9059cbb000000000000000000000000742d35cc6634c0532925a3b844bc454e4438f44e0000000000000000000000000000000000000000000000000de0b6b3a7640000",
    "chain": "ethereum",
    "status": "pending",
    "block_number": 0,
    "block_hash": "",
    "confirmations": 0,
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### Get Token Information

```
GET /api/v1/contracts/token-info
```

**Query Parameters:**

- `address` (required): Token contract address
- `chain` (required): Blockchain (ethereum, bsc, polygon)

**Response:**

```json
{
  "name": "My Token",
  "symbol": "MTK",
  "decimals": 18,
  "total_supply": "1000000000000000000000000",
  "type": "erc20"
}
```

## Security API

### Generate a Key Pair

```
POST /api/v1/security/generate-key-pair
```

**Request Body:**

```json
{
  "chain": "ethereum"
}
```

**Response:**

```json
{
  "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "public_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
}
```

### Generate a Mnemonic

```
POST /api/v1/security/generate-mnemonic
```

**Request Body:**

```json
{
  "strength": 256
}
```

**Response:**

```json
{
  "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
}
```

### Validate a Mnemonic

```
POST /api/v1/security/validate-mnemonic
```

**Request Body:**

```json
{
  "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
}
```

**Response:**

```json
{
  "valid": true
}
```

### Convert Mnemonic to Private Key

```
POST /api/v1/security/mnemonic-to-private-key
```

**Request Body:**

```json
{
  "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
  "path": "m/44'/60'/0'/0/0"
}
```

**Response:**

```json
{
  "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "public_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
}
```

## Error Handling

All API endpoints return standard HTTP status codes. In case of an error, the response body will contain an error message.

**Example Error Response:**

```json
{
  "error": "Wallet not found",
  "code": "not_found"
}
```

### Common Error Codes

- `bad_request`: The request was invalid (HTTP 400)
- `unauthorized`: Authentication is required (HTTP 401)
- `forbidden`: The user does not have permission (HTTP 403)
- `not_found`: The requested resource was not found (HTTP 404)
- `conflict`: The request conflicts with the current state (HTTP 409)
- `internal_server_error`: An internal server error occurred (HTTP 500)
