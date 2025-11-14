package model

// CreateOrderRequest represents the HTTP request body for creating an order
// For Hosted Checkout (redirect), PayOptionDetails is NOT required
type CreateOrderRequest struct {
	PartnerReferenceNo string                   `json:"partner_reference_no" binding:"required"`
	MerchantID         string                   `json:"merchant_id,omitempty"`
	Amount             MoneyRequest             `json:"amount" binding:"required"`
	PayOptionDetails   []PayOptionDetailRequest `json:"pay_option_details,omitempty"` // Optional for hosted checkout
	UrlParams          []UrlParamRequest        `json:"url_params" binding:"required"`
	SubMerchantID      *string                  `json:"sub_merchant_id,omitempty"`
	ExternalStoreID    *string                  `json:"external_store_id,omitempty"`
	ValidUpTo          *string                  `json:"valid_up_to,omitempty"`
	DisabledPayMethods *string                  `json:"disabled_pay_methods,omitempty"`
}

// MoneyRequest represents amount and currency
type MoneyRequest struct {
	Value    string `json:"value" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

// PayOptionDetailRequest represents payment option details
type PayOptionDetailRequest struct {
	PayMethod     string        `json:"pay_method" binding:"required"`
	PayOption     string        `json:"pay_option" binding:"required"`
	TransAmount   MoneyRequest  `json:"trans_amount" binding:"required"`
	FeeAmount     *MoneyRequest `json:"fee_amount,omitempty"`
	CardToken     *string       `json:"card_token,omitempty"`
	MerchantToken *string       `json:"merchant_token,omitempty"`
}

// UrlParamRequest represents URL parameters for notifications
type UrlParamRequest struct {
	Url        string `json:"url" binding:"required"`
	Type       string `json:"type" binding:"required"`        // PAY_RETURN or NOTIFICATION
	IsDeeplink string `json:"is_deeplink" binding:"required"` // "true" or "false"
}
