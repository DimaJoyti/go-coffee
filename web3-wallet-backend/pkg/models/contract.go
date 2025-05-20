package models

import (
	"time"
)

// Contract represents a smart contract
type Contract struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	Chain     Chain     `json:"chain" db:"chain"`
	ABI       string    `json:"abi" db:"abi"`
	Bytecode  string    `json:"bytecode" db:"bytecode"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ContractType represents the type of contract
type ContractType string

const (
	// ContractTypeERC20 represents an ERC-20 token contract
	ContractTypeERC20 ContractType = "erc20"
	// ContractTypeERC721 represents an ERC-721 token contract (NFT)
	ContractTypeERC721 ContractType = "erc721"
	// ContractTypeERC1155 represents an ERC-1155 token contract (Multi-token)
	ContractTypeERC1155 ContractType = "erc1155"
	// ContractTypeCustom represents a custom contract
	ContractTypeCustom ContractType = "custom"
)

// DeployContractRequest represents a request to deploy a contract
type DeployContractRequest struct {
	UserID    string       `json:"user_id" validate:"required"`
	WalletID  string       `json:"wallet_id" validate:"required"`
	Name      string       `json:"name" validate:"required"`
	Chain     Chain        `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
	Type      ContractType `json:"type" validate:"required,oneof=erc20 erc721 erc1155 custom"`
	ABI       string       `json:"abi" validate:"required"`
	Bytecode  string       `json:"bytecode" validate:"required"`
	Arguments []string     `json:"arguments"`
	Gas       uint64       `json:"gas"`
	GasPrice  string       `json:"gas_price"`
	Passphrase string      `json:"passphrase" validate:"required"`
}

// DeployContractResponse represents a response to a deploy contract request
type DeployContractResponse struct {
	Contract    Contract    `json:"contract"`
	Transaction Transaction `json:"transaction"`
}

// ImportContractRequest represents a request to import a contract
type ImportContractRequest struct {
	UserID  string       `json:"user_id" validate:"required"`
	Name    string       `json:"name" validate:"required"`
	Address string       `json:"address" validate:"required"`
	Chain   Chain        `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
	Type    ContractType `json:"type" validate:"required,oneof=erc20 erc721 erc1155 custom"`
	ABI     string       `json:"abi" validate:"required"`
}

// ImportContractResponse represents a response to an import contract request
type ImportContractResponse struct {
	Contract Contract `json:"contract"`
}

// GetContractRequest represents a request to get a contract
type GetContractRequest struct {
	ID string `json:"id" validate:"required"`
}

// GetContractResponse represents a response to a get contract request
type GetContractResponse struct {
	Contract Contract `json:"contract"`
}

// GetContractByAddressRequest represents a request to get a contract by address
type GetContractByAddressRequest struct {
	Address string `json:"address" validate:"required"`
	Chain   Chain  `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
}

// GetContractByAddressResponse represents a response to a get contract by address request
type GetContractByAddressResponse struct {
	Contract Contract `json:"contract"`
}

// ListContractsRequest represents a request to list contracts
type ListContractsRequest struct {
	UserID string       `json:"user_id" validate:"required"`
	Chain  Chain        `json:"chain"`
	Type   ContractType `json:"type"`
	Limit  int          `json:"limit" validate:"min=1,max=100"`
	Offset int          `json:"offset" validate:"min=0"`
}

// ListContractsResponse represents a response to a list contracts request
type ListContractsResponse struct {
	Contracts []Contract `json:"contracts"`
	Total     int        `json:"total"`
}

// CallContractRequest represents a request to call a contract method
type CallContractRequest struct {
	ContractID string   `json:"contract_id" validate:"required"`
	Method     string   `json:"method" validate:"required"`
	Arguments  []string `json:"arguments"`
	From       string   `json:"from"`
}

// CallContractResponse represents a response to a call contract method request
type CallContractResponse struct {
	Result interface{} `json:"result"`
}

// SendContractTransactionRequest represents a request to send a contract transaction
type SendContractTransactionRequest struct {
	ContractID string   `json:"contract_id" validate:"required"`
	WalletID   string   `json:"wallet_id" validate:"required"`
	Method     string   `json:"method" validate:"required"`
	Arguments  []string `json:"arguments"`
	Value      string   `json:"value"`
	Gas        uint64   `json:"gas"`
	GasPrice   string   `json:"gas_price"`
	Passphrase string   `json:"passphrase" validate:"required"`
}

// SendContractTransactionResponse represents a response to a send contract transaction request
type SendContractTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// GetContractEventsRequest represents a request to get contract events
type GetContractEventsRequest struct {
	ContractID string `json:"contract_id" validate:"required"`
	Event      string `json:"event" validate:"required"`
	FromBlock  uint64 `json:"from_block"`
	ToBlock    uint64 `json:"to_block"`
	Limit      int    `json:"limit" validate:"min=1,max=100"`
	Offset     int    `json:"offset" validate:"min=0"`
}

// GetContractEventsResponse represents a response to a get contract events request
type GetContractEventsResponse struct {
	Events []ContractEvent `json:"events"`
	Total  int             `json:"total"`
}

// ContractEvent represents a contract event
type ContractEvent struct {
	ContractID    string                 `json:"contract_id"`
	Event         string                 `json:"event"`
	TransactionID string                 `json:"transaction_id"`
	BlockNumber   uint64                 `json:"block_number"`
	LogIndex      uint                   `json:"log_index"`
	Data          map[string]interface{} `json:"data"`
	CreatedAt     time.Time              `json:"created_at"`
}

// GetTokenInfoRequest represents a request to get token information
type GetTokenInfoRequest struct {
	Address string `json:"address" validate:"required"`
	Chain   Chain  `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
}

// GetTokenInfoResponse represents a response to a get token information request
type GetTokenInfoResponse struct {
	Name     string       `json:"name"`
	Symbol   string       `json:"symbol"`
	Decimals int          `json:"decimals"`
	TotalSupply string    `json:"total_supply"`
	Type     ContractType `json:"type"`
}
