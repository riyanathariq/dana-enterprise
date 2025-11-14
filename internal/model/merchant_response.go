package model

// MerchantInfoResponse represents a clean merchant information response
type MerchantInfoResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    *MerchantInfoData `json:"data"`
	Meta    *MerchantInfoMeta `json:"meta,omitempty"`
}

// MerchantInfoData contains the actual merchant information
type MerchantInfoData struct {
	MerchantID string                       `json:"merchant_id"`
	Balances   *MerchantBalances            `json:"balances"`
	Resources  map[string]*MerchantResource `json:"resources"`
}

// MerchantBalances contains all balance information
type MerchantBalances struct {
	DepositBalance   *BalanceInfo `json:"deposit_balance,omitempty"`
	AvailableBalance *BalanceInfo `json:"available_balance,omitempty"`
	TotalBalance     *BalanceInfo `json:"total_balance,omitempty"`
}

// BalanceInfo contains balance details
type BalanceInfo struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// MerchantResource contains resource information
type MerchantResource struct {
	Type        string `json:"type"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// MerchantInfoMeta contains metadata about the response
type MerchantInfoMeta struct {
	RequestID   string `json:"request_id,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
