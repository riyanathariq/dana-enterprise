package order

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dana-id/dana-go/payment_gateway/v1"
	danaSDK "github.com/riyanathariq/dana-enterprise/internal/sdk/dana"
	"github.com/riyanathariq/dana-enterprise/package/dana"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// formatAmountValue ensures amount value has exactly 2 decimal places for IDR currency
// DANA API requires format: "100000.00" not "100000" or "100000.0"
func formatAmountValue(value string) string {
	// If value is empty, return as is
	if value == "" {
		return value
	}

	// Remove any whitespace
	value = strings.TrimSpace(value)

	// Check if value already has decimal point
	if strings.Contains(value, ".") {
		parts := strings.Split(value, ".")
		intPart := parts[0]
		decimalPart := parts[1]

		// Ensure decimal part has exactly 2 digits
		if len(decimalPart) == 0 {
			decimalPart = "00"
		} else if len(decimalPart) == 1 {
			decimalPart = decimalPart + "0"
		} else if len(decimalPart) > 2 {
			// Truncate to 2 decimal places
			decimalPart = decimalPart[:2]
		}

		return intPart + "." + decimalPart
	}

	// If no decimal point, add ".00"
	return value + ".00"
}

// CreateOrderRequestParams contains parameters for creating an order
type CreateOrderRequestParams struct {
	PartnerReferenceNo string                            // Required: Transaction identifier on partner system
	MerchantID         string                            // Required: Merchant identifier
	Amount             payment_gateway.Money             // Required: Amount with currency
	PayOptionDetails   []payment_gateway.PayOptionDetail // Required for custom checkout, optional for hosted checkout
	UrlParams          []payment_gateway.UrlParam        // Required: Notification URLs
	SubMerchantID      *string                           // Optional: Sub merchant identifier
	ExternalStoreID    *string                           // Optional: Store identifier
	ValidUpTo          *string                           // Optional: Expiration time (YYYY-MM-DDTHH:mm:ss+07:00)
	DisabledPayMethods *string                           // Optional: Disabled payment methods
}

// normalizeUrlParams normalizes URL parameters for both checkout types
func normalizeUrlParams(params []payment_gateway.UrlParam) ([]payment_gateway.UrlParam, error) {
	normalizedUrlParams := make([]payment_gateway.UrlParam, len(params))
	for i, up := range params {
		isDeeplink := strings.ToUpper(strings.TrimSpace(up.IsDeeplink))
		// Normalize: "Y" -> "Y", "N" -> "N", "true" -> "Y", "false" -> "N"
		if isDeeplink == "TRUE" {
			isDeeplink = "Y"
		} else if isDeeplink == "FALSE" {
			isDeeplink = "N"
		} else if isDeeplink != "Y" && isDeeplink != "N" {
			// Default to "N" if invalid
			isDeeplink = "N"
		}

		// NOTIFICATION type should always be "N" (not deeplink) as it's server-to-server webhook
		if strings.ToUpper(up.Type) == "NOTIFICATION" {
			isDeeplink = "N"
		}

		// Validate URL format
		url := strings.TrimSpace(up.Url)
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return nil, fmt.Errorf("invalid URL format for urlParams[%d]: must start with http:// or https://", i)
		}

		normalizedUrlParams[i] = payment_gateway.UrlParam{
			Url:        url,
			Type:       up.Type,
			IsDeeplink: isDeeplink,
		}
	}
	return normalizedUrlParams, nil
}

// formatValidUpTo formats validUpTo in Jakarta timezone
func formatValidUpTo(validUpTo *string) *string {
	if validUpTo != nil && *validUpTo != "" {
		return validUpTo
	}
	jakartaTz, _ := time.LoadLocation("Asia/Jakarta")
	expiredTime := time.Now().In(jakartaTz).Add(1 * time.Hour).Format("2006-01-02T15:04:05+07:00")
	return &expiredTime
}

// CreateOrderHostedCheckout creates a new order using Hosted Checkout (Redirect)
// User will be redirected to DANA payment page to select payment method
// Uses raw HTTP request instead of SDK
func (s *Service) CreateOrderHostedCheckout(ctx context.Context, params CreateOrderRequestParams) (*payment_gateway.CreateOrderResponse, error) {
	// Use merchant ID from params or fallback to env
	merchantID := params.MerchantID
	if merchantID == "" {
		merchantID = os.Getenv("DANA_MERCHANT_ID")
	}

	// Validate required fields
	if params.PartnerReferenceNo == "" {
		return nil, fmt.Errorf("partnerReferenceNo is required")
	}
	if merchantID == "" {
		return nil, fmt.Errorf("merchantId is required")
	}
	if len(params.UrlParams) == 0 {
		return nil, fmt.Errorf("urlParams is required and cannot be empty")
	}

	// Format amount value to ensure 2 decimal places for IDR
	formattedAmount := params.Amount
	if params.Amount.Currency == "IDR" {
		formattedAmount.Value = formatAmountValue(params.Amount.Value)
	}

	validUpTo := formatValidUpTo(params.ValidUpTo)
	normalizedUrlParams, err := normalizeUrlParams(params.UrlParams)
	if err != nil {
		return nil, err
	}

	// AdditionalInfo for Hosted Checkout (redirect)
	// Use MCC code from environment or default to a common one
	// Common MCC codes: 5411 (Grocery), 5999 (Miscellaneous), 5812 (Restaurants)
	mccCode := os.Getenv("DANA_MCC")
	if mccCode == "" {
		mccCode = "5999" // Default to Miscellaneous if not set
	}
	webTerminalType := "WEB"

	// Order object - required for hosted checkout redirect scenario
	orderTitle := os.Getenv("DANA_ORDER_TITLE")
	if orderTitle == "" {
		orderTitle = "Order " + params.PartnerReferenceNo // Default order title
	}
	scenario := "REDIRECT" // Required for hosted checkout redirect
	orderObj := &payment_gateway.OrderRedirectObject{
		OrderTitle: orderTitle,
		Scenario:   &scenario,
	}

	additionalInfo := &payment_gateway.CreateOrderByRedirectAdditionalInfo{
		Mcc: mccCode,
		EnvInfo: payment_gateway.EnvInfo{
			SourcePlatform:    "IPG",
			TerminalType:      webTerminalType,
			OrderTerminalType: &webTerminalType,
		},
		Order: orderObj,
	}

	// Use raw HTTP request instead of SDK
	// Convert to raw SDK params
	rawParams := danaSDK.CreateOrderRequestParams{
		PartnerReferenceNo: params.PartnerReferenceNo,
		MerchantID:         merchantID,
		Amount:             formattedAmount,
		UrlParams:          normalizedUrlParams,
		SubMerchantID:      params.SubMerchantID,
		ExternalStoreID:    params.ExternalStoreID,
		ValidUpTo:          validUpTo,
		DisabledPayMethods: params.DisabledPayMethods,
		AdditionalInfo:     additionalInfo,
	}

	// Call raw HTTP request for hosted checkout
	rawResponse, err := danaSDK.CreateOrderHostedRaw(ctx, rawParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create order (raw HTTP): %w", err)
	}

	// Convert raw response to SDK response format
	// The raw response structure matches DANA API response
	order := &payment_gateway.CreateOrderResponse{
		ResponseCode:       rawResponse.ResponseCode,
		ResponseMessage:    rawResponse.ResponseMessage,
		PartnerReferenceNo: rawResponse.PartnerReferenceNo,
	}

	// If response has data, try to unmarshal it to the full response structure
	if rawResponse.Data != nil {
		if dataBytes, err := json.Marshal(rawResponse.Data); err == nil {
			// Try to unmarshal into the full CreateOrderResponse structure
			var fullResponse payment_gateway.CreateOrderResponse
			if err := json.Unmarshal(dataBytes, &fullResponse); err == nil {
				order = &fullResponse
			} else {
				// If unmarshal fails, at least we have the basic fields
				fmt.Printf("Warning: Could not unmarshal response data: %v\n", err)
			}
		}
	}

	return order, nil
}

// CreateOrderCustomCheckout creates a new order using Custom Checkout (Host-to-Host)
// Requires PayOptionDetails to specify payment method directly
// Uses raw HTTP request instead of SDK
func (s *Service) CreateOrderCustomCheckout(ctx context.Context, params CreateOrderRequestParams) (*payment_gateway.CreateOrderResponse, error) {

	// Use merchant ID from params or fallback to env
	merchantID := params.MerchantID
	if merchantID == "" {
		merchantID = os.Getenv("DANA_MERCHANT_ID")
	}

	// Validate required fields
	if params.PartnerReferenceNo == "" {
		return nil, fmt.Errorf("partnerReferenceNo is required")
	}
	if merchantID == "" {
		return nil, fmt.Errorf("merchantId is required")
	}
	if len(params.PayOptionDetails) == 0 {
		return nil, fmt.Errorf("payOptionDetails is required for custom checkout")
	}
	if len(params.UrlParams) == 0 {
		return nil, fmt.Errorf("urlParams is required and cannot be empty")
	}

	// Format amount value to ensure 2 decimal places for IDR
	formattedAmount := params.Amount
	if params.Amount.Currency == "IDR" {
		formattedAmount.Value = formatAmountValue(params.Amount.Value)
	}

	// Format PayOptionDetails amounts
	formattedPayOptionDetails := make([]payment_gateway.PayOptionDetail, len(params.PayOptionDetails))
	totalTransAmount := 0.0
	for i, pod := range params.PayOptionDetails {
		formattedPayOptionDetails[i] = pod
		if pod.TransAmount.Currency == "IDR" {
			formattedPayOptionDetails[i].TransAmount.Value = formatAmountValue(pod.TransAmount.Value)
			// Track total for validation
			if val, err := strconv.ParseFloat(formattedPayOptionDetails[i].TransAmount.Value, 64); err == nil {
				totalTransAmount += val
			}
		}
		if pod.FeeAmount != nil && pod.FeeAmount.Currency == "IDR" {
			feeValue := formatAmountValue(pod.FeeAmount.Value)
			formattedFeeAmount := &payment_gateway.Money{
				Value:    feeValue,
				Currency: pod.FeeAmount.Currency,
			}
			formattedPayOptionDetails[i].FeeAmount = formattedFeeAmount
		}
	}

	// Validate: total transAmount should match main amount (for IDR)
	if formattedAmount.Currency == "IDR" {
		if mainAmount, err := strconv.ParseFloat(formattedAmount.Value, 64); err == nil {
			if totalTransAmount > 0 && mainAmount != totalTransAmount {
				// Warning: total transAmount doesn't match main amount, but continue anyway
				if debug, _ := strconv.ParseBool(os.Getenv("DANA_DEBUG")); debug {
					fmt.Printf("WARNING: Total transAmount (%.2f) doesn't match main amount (%.2f)\n", totalTransAmount, mainAmount)
				}
			}
		}
	}

	validUpTo := formatValidUpTo(params.ValidUpTo)
	normalizedUrlParams, err := normalizeUrlParams(params.UrlParams)
	if err != nil {
		return nil, err
	}

	// AdditionalInfo for Custom Checkout (Host-to-Host)
	mccCode := os.Getenv("DANA_MCC")
	if mccCode == "" {
		mccCode = "5999" // Default to Miscellaneous if not set
	}
	webTerminalType := "WEB"

	// Order object - required for custom checkout
	// Based on DANA documentation, need orderTitle, scenario, merchantTransType, and buyer (REQUIRED)
	orderTitle := os.Getenv("DANA_ORDER_TITLE")
	if orderTitle == "" {
		orderTitle = "Order " + params.PartnerReferenceNo
	}
	scenario := "API" // Required for custom checkout (host-to-host)
	merchantTransType := os.Getenv("DANA_MERCHANT_TRANS_TYPE")
	if merchantTransType == "" {
		merchantTransType = "SALE" // Default to SALE if not set
	}

	// Buyer is REQUIRED for custom checkout
	// All fields are optional, but buyer object itself is required
	buyerObj := &payment_gateway.Buyer{}
	if externalUserId := os.Getenv("DANA_BUYER_EXTERNAL_USER_ID"); externalUserId != "" {
		buyerObj.ExternalUserId = &externalUserId
	}
	if userId := os.Getenv("DANA_BUYER_USER_ID"); userId != "" {
		buyerObj.UserId = &userId
	}
	if nickname := os.Getenv("DANA_BUYER_NICKNAME"); nickname != "" {
		buyerObj.Nickname = &nickname
	}
	if externalUserType := os.Getenv("DANA_BUYER_EXTERNAL_USER_TYPE"); externalUserType != "" {
		buyerObj.ExternalUserType = &externalUserType
	}

	orderObj := &payment_gateway.OrderApiObject{
		OrderTitle:        orderTitle,
		Scenario:          &scenario,
		MerchantTransType: &merchantTransType,
		Buyer:             buyerObj, // REQUIRED
		// goods, shippingInfo are optional
	}

	// EnvInfo for custom checkout - add more fields if needed
	envInfo := payment_gateway.EnvInfo{
		SourcePlatform:    "IPG",
		TerminalType:      webTerminalType,
		OrderTerminalType: &webTerminalType,
	}

	// Optional: Add more envInfo fields if provided
	if clientIp := os.Getenv("DANA_CLIENT_IP"); clientIp != "" {
		envInfo.ClientIp = &clientIp
	}
	if sessionId := os.Getenv("DANA_SESSION_ID"); sessionId != "" {
		envInfo.SessionId = &sessionId
	}
	if tokenId := os.Getenv("DANA_TOKEN_ID"); tokenId != "" {
		envInfo.TokenId = &tokenId
	}
	if osType := os.Getenv("DANA_OS_TYPE"); osType != "" {
		envInfo.OsType = &osType
	}
	if websiteLanguage := os.Getenv("DANA_WEBSITE_LANGUAGE"); websiteLanguage != "" {
		envInfo.WebsiteLanguage = &websiteLanguage
	}

	additionalInfo := &payment_gateway.CreateOrderByApiAdditionalInfo{
		Mcc:     mccCode,
		EnvInfo: envInfo,
		Order:   orderObj,
	}

	// Use raw HTTP request instead of SDK
	// Convert to raw SDK params
	rawParams := danaSDK.CreateOrderRequestParams{
		PartnerReferenceNo: params.PartnerReferenceNo,
		MerchantID:         merchantID,
		Amount:             formattedAmount,
		PayOptionDetails:   formattedPayOptionDetails,
		UrlParams:          normalizedUrlParams,
		SubMerchantID:      params.SubMerchantID,
		ExternalStoreID:    params.ExternalStoreID,
		ValidUpTo:          validUpTo,
		DisabledPayMethods: params.DisabledPayMethods,
		AdditionalInfo:     additionalInfo,
	}

	// Call raw HTTP request
	rawResponse, err := danaSDK.CreateOrderRaw(ctx, rawParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create order (raw HTTP): %w", err)
	}

	// Convert raw response to SDK response format
	// The raw response structure matches DANA API response
	order := &payment_gateway.CreateOrderResponse{
		ResponseCode:       rawResponse.ResponseCode,
		ResponseMessage:    rawResponse.ResponseMessage,
		PartnerReferenceNo: rawResponse.PartnerReferenceNo,
	}

	// If response has data, try to unmarshal it to the full response structure
	if rawResponse.Data != nil {
		if dataBytes, err := json.Marshal(rawResponse.Data); err == nil {
			// Try to unmarshal into the full CreateOrderResponse structure
			var fullResponse payment_gateway.CreateOrderResponse
			if err := json.Unmarshal(dataBytes, &fullResponse); err == nil {
				order = &fullResponse
			} else {
				// If unmarshal fails, at least we have the basic fields
				fmt.Printf("Warning: Could not unmarshal response data: %v\n", err)
			}
		}
	}

	return order, nil
}

func (s *Service) GetPaymentMethod(ctx context.Context) (*payment_gateway.ConsultPayResponse, error) {
	danaClient := dana.InitData()

	_ = os.Getenv("DANA_MERCHANT_ID")

	webTerminalType := "WEB"

	// Format amount value to ensure 2 decimal places
	amountValue := formatAmountValue("100000")
	paymentMethod, _, err := danaClient.PaymentGatewayAPI.ConsultPay(ctx).
		ConsultPayRequest(payment_gateway.ConsultPayRequest{
			MerchantId: "",
			Amount: payment_gateway.Money{
				Value:    amountValue,
				Currency: "IDR",
			},
			AdditionalInfo: payment_gateway.ConsultPayRequestAdditionalInfo{
				Buyer: payment_gateway.Buyer{},
				EnvInfo: payment_gateway.EnvInfo{
					SourcePlatform:    "IPG",
					TerminalType:      webTerminalType,
					OrderTerminalType: &webTerminalType,
				},
			},
		}).
		Execute()
	if err != nil {
		return nil, err
	}
	return paymentMethod, nil
}

func (s *Service) GetOrder(ctx context.Context, partnerReferenceNo string) (*payment_gateway.QueryPaymentResponse, error) {
	danaClient := dana.InitData()

	merchantID := os.Getenv("DANA_MERCHANT_ID")

	order, _, err := danaClient.PaymentGatewayAPI.QueryPayment(ctx).
		QueryPaymentRequest(payment_gateway.QueryPaymentRequest{
			OriginalPartnerReferenceNo: &partnerReferenceNo,
			MerchantId:                 merchantID,
			ServiceCode:                "54",
		}).
		Execute()
	if err != nil {
		return nil, err
	}
	return order, nil
}
