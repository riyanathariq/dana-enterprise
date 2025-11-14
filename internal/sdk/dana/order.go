package dana

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dana-id/dana-go/payment_gateway/v1"
	uuid "github.com/google/uuid"
)

// CreateOrderRequestParams represents the parameters for creating an order
type CreateOrderRequestParams struct {
	PartnerReferenceNo string
	MerchantID         string
	Amount             payment_gateway.Money
	PayOptionDetails   []payment_gateway.PayOptionDetail
	UrlParams          []payment_gateway.UrlParam
	SubMerchantID      *string
	ExternalStoreID    *string
	ValidUpTo          *string
	DisabledPayMethods *string
	AdditionalInfo     interface{}
}

// CreateOrderResponse represents the response from DANA API
type CreateOrderResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	Data               interface{} `json:"data,omitempty"`
}

// getEnv returns environment variable value or default if empty
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// CreateOrderRaw creates an order using raw HTTP request without SDK
func CreateOrderRaw(ctx context.Context, params CreateOrderRequestParams) (*CreateOrderResponse, error) {
	// Get environment variables
	env := getEnv("DANA_ENV", "sandbox")
	var baseURL string
	if env == "production" {
		baseURL = "https://api.dana.id"
	} else {
		if host := getEnv("DANA_HOST", ""); host != "" {
			scheme := getEnv("DANA_SCHEME", "https")
			baseURL = scheme + "://" + host
		} else {
			baseURL = "https://api.sandbox.dana.id"
		}
	}

	// Build request body
	requestBody := map[string]interface{}{
		"partnerReferenceNo": params.PartnerReferenceNo,
		"merchantId":         params.MerchantID,
		"amount": map[string]string{
			"value":    params.Amount.Value,
			"currency": params.Amount.Currency,
		},
	}

	// Add PayOptionDetails if provided
	if len(params.PayOptionDetails) > 0 {
		payOptionDetails := make([]map[string]interface{}, len(params.PayOptionDetails))
		for i, pod := range params.PayOptionDetails {
			podMap := map[string]interface{}{
				"payMethod": pod.PayMethod,
				"payOption": pod.PayOption,
				"transAmount": map[string]string{
					"value":    pod.TransAmount.Value,
					"currency": pod.TransAmount.Currency,
				},
			}
			if pod.FeeAmount != nil {
				podMap["feeAmount"] = map[string]string{
					"value":    pod.FeeAmount.Value,
					"currency": pod.FeeAmount.Currency,
				}
			}
			if pod.CardToken != nil {
				podMap["cardToken"] = *pod.CardToken
			}
			if pod.MerchantToken != nil {
				podMap["merchantToken"] = *pod.MerchantToken
			}
			payOptionDetails[i] = podMap
		}
		requestBody["payOptionDetails"] = payOptionDetails
	}

	// Add UrlParams
	urlParams := make([]map[string]string, len(params.UrlParams))
	for i, up := range params.UrlParams {
		urlParams[i] = map[string]string{
			"url":        up.Url,
			"type":       up.Type,
			"isDeeplink": up.IsDeeplink,
		}
	}
	requestBody["urlParams"] = urlParams

	// Add optional fields
	if params.SubMerchantID != nil {
		requestBody["subMerchantId"] = *params.SubMerchantID
	}
	if params.ExternalStoreID != nil {
		requestBody["externalStoreId"] = *params.ExternalStoreID
	}
	if params.ValidUpTo != nil {
		requestBody["validUpTo"] = *params.ValidUpTo
	}
	if params.DisabledPayMethods != nil {
		requestBody["disabledPayMethods"] = *params.DisabledPayMethods
	}
	if params.AdditionalInfo != nil {
		requestBody["additionalInfo"] = params.AdditionalInfo
	}

	// Marshal request body to JSON (minified, no indentation) - same as SDK
	// First marshal to JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Compact/minify JSON like SDK does (json.Compact removes whitespace)
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, bodyBytes); err != nil {
		return nil, fmt.Errorf("failed to compact JSON: %w", err)
	}
	bodyBytes = compacted.Bytes()
	bodyString := string(bodyBytes)

	// Get Jakarta timezone timestamp
	jkt, err := time.LoadLocation("Asia/Jakarta")
	var jktTime time.Time
	if err != nil {
		now := time.Now().UTC()
		jktTime = now.Add(7 * time.Hour)
	} else {
		jktTime = time.Now().In(jkt)
	}
	timestamp := jktTime.Format("2006-01-02T15:04:05+07:00")

	// Build endpoint URL - same as SDK
	endpoint := baseURL + "/payment-gateway/v1.0/debit/payment-host-to-host.htm"
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}
	urlPath := endpointURL.Path

	// Generate signature
	// Format: "<HTTP METHOD>:<RELATIVE PATH URL>:<LOWERCASE_HEX_ENCODED_SHA_256(MINIFIED_HTTP_BODY)>:<X-TIMESTAMP>"
	hash := sha256.New()
	hash.Write(bodyBytes)
	hashedPayload := fmt.Sprintf("%x", hash.Sum(nil))

	stringToSign := fmt.Sprintf("POST:%s:%s:%s", urlPath, hashedPayload, timestamp)

	// Sign with private key
	privateKeyStr := os.Getenv("DANA_PRIVATE_KEY")
	if privateKeyStr == "" {
		return nil, fmt.Errorf("DANA_PRIVATE_KEY is required")
	}

	// Normalize private key (handle \n literals)
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "\\n", "\n")
	if !strings.Contains(privateKeyStr, "-----BEGIN") {
		return nil, fmt.Errorf("invalid private key format: missing PEM headers")
	}

	privateKeyPEM := []byte(privateKeyStr)
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}

	var privateKey *rsa.PrivateKey
	// Try PKCS1 first
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parsed key is not an RSA private key")
		}
		privateKey = rsaKey
	}

	// Create SHA-256 hash of the string to sign
	hash = sha256.New()
	hash.Write([]byte(stringToSign))
	hashed := hash.Sum(nil)

	// Sign the hashed data with PKCS1v15
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode signature as base64
	signature := base64.StdEncoding.EncodeToString(signatureBytes)

	// Get partner ID
	partnerID := getEnv("DANA_X_PARTNER_ID", "")
	if partnerID == "" {
		partnerID = os.Getenv("DANA_CLIENT_ID")
	}

	// Generate external ID
	externalID := "sdk" + uuid.New().String()[3:]

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-SIGNATURE", signature)
	req.Header.Set("X-PARTNER-ID", partnerID)
	req.Header.Set("X-EXTERNAL-ID", externalID)
	req.Header.Set("CHANNEL-ID", "95221")

	// Set optional headers
	if origin := getEnv("DANA_ORIGIN", ""); origin != "" {
		req.Header.Set("ORIGIN", origin)
	}

	// Set debug mode if enabled
	if debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false")); debug {
		if strings.ToLower(env) == "sandbox" {
			req.Header.Set("X-Debug-Mode", "true")
		}
	}

	// Debug: Print request details
	if debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false")); debug {
		fmt.Printf("DEBUG: Raw HTTP Request:\n")
		fmt.Printf("  URL: %s\n", endpoint)
		fmt.Printf("  Method: POST\n")
		fmt.Printf("  Headers:\n")
		for k, v := range req.Header {
			if k == "X-SIGNATURE" {
				fmt.Printf("    %s: %s (truncated)\n", k, v[0][:20]+"...")
			} else {
				fmt.Printf("    %s: %s\n", k, v[0])
			}
		}
		fmt.Printf("  Body:\n%s\n", bodyString)
		fmt.Printf("  String to Sign: %s\n", stringToSign)
	}

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug: Print response
	if debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false")); debug {
		fmt.Printf("DEBUG: Raw HTTP Response:\n")
		fmt.Printf("  Status: %s\n", resp.Status)
		fmt.Printf("  StatusCode: %d\n", resp.StatusCode)
		fmt.Printf("  Body:\n%s\n", string(respBody))
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errData interface{}
		if json.Unmarshal(respBody, &errData) == nil {
			errJSON, _ := json.MarshalIndent(errData, "", "  ")
			return nil, fmt.Errorf("DANA API error (HTTP %d): %s", resp.StatusCode, string(errJSON))
		}
		return nil, fmt.Errorf("DANA API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var response CreateOrderResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// CreateOrderHostedRaw creates an order using Hosted Checkout (Redirect) with raw HTTP request without SDK
// User will be redirected to DANA payment page to select payment method
func CreateOrderHostedRaw(ctx context.Context, params CreateOrderRequestParams) (*CreateOrderResponse, error) {
	// Get environment variables
	env := getEnv("DANA_ENV", "sandbox")
	var baseURL string
	if env == "production" {
		baseURL = "https://api.dana.id"
	} else {
		if host := getEnv("DANA_HOST", ""); host != "" {
			scheme := getEnv("DANA_SCHEME", "https")
			baseURL = scheme + "://" + host
		} else {
			baseURL = "https://api.sandbox.dana.id"
		}
	}

	// Build request body for Hosted Checkout (NO payOptionDetails)
	requestBody := map[string]interface{}{
		"partnerReferenceNo": params.PartnerReferenceNo,
		"merchantId":         params.MerchantID,
		"amount": map[string]string{
			"value":    params.Amount.Value,
			"currency": params.Amount.Currency,
		},
	}

	// Add UrlParams
	urlParams := make([]map[string]string, len(params.UrlParams))
	for i, up := range params.UrlParams {
		urlParams[i] = map[string]string{
			"url":        up.Url,
			"type":       up.Type,
			"isDeeplink": up.IsDeeplink,
		}
	}
	requestBody["urlParams"] = urlParams

	// Add optional fields
	if params.SubMerchantID != nil {
		requestBody["subMerchantId"] = *params.SubMerchantID
	}
	if params.ExternalStoreID != nil {
		requestBody["externalStoreId"] = *params.ExternalStoreID
	}
	if params.ValidUpTo != nil {
		requestBody["validUpTo"] = *params.ValidUpTo
	}
	if params.DisabledPayMethods != nil {
		requestBody["disabledPayMethods"] = *params.DisabledPayMethods
	}

	// Build AdditionalInfo with Order object for Hosted Checkout
	// Use MCC code from environment or default
	mccCode := getEnv("DANA_MCC", "5999")
	webTerminalType := "WEB"

	// Order object - required for hosted checkout redirect scenario
	orderTitle := getEnv("DANA_ORDER_TITLE", "")
	if orderTitle == "" {
		orderTitle = "Order " + params.PartnerReferenceNo
	}
	scenario := "REDIRECT" // Required for hosted checkout redirect

	// Build additionalInfo with order object
	additionalInfo := map[string]interface{}{
		"mcc": mccCode,
		"envInfo": map[string]interface{}{
			"sourcePlatform":    "IPG",
			"terminalType":      webTerminalType,
			"orderTerminalType": webTerminalType,
		},
		"order": map[string]interface{}{
			"orderTitle": orderTitle,
			"scenario":   scenario,
		},
	}

	// Override with params.AdditionalInfo if provided, otherwise use built one
	if params.AdditionalInfo != nil {
		requestBody["additionalInfo"] = params.AdditionalInfo
	} else {
		requestBody["additionalInfo"] = additionalInfo
	}

	// Marshal request body to JSON (minified, no indentation) - same as SDK
	// First marshal to JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Compact/minify JSON like SDK does (json.Compact removes whitespace)
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, bodyBytes); err != nil {
		return nil, fmt.Errorf("failed to compact JSON: %w", err)
	}
	bodyBytes = compacted.Bytes()
	bodyString := string(bodyBytes)

	// Get Jakarta timezone timestamp
	jkt, err := time.LoadLocation("Asia/Jakarta")
	var jktTime time.Time
	if err != nil {
		now := time.Now().UTC()
		jktTime = now.Add(7 * time.Hour)
	} else {
		jktTime = time.Now().In(jkt)
	}
	timestamp := jktTime.Format("2006-01-02T15:04:05+07:00")

	// Build endpoint URL - same as SDK (same endpoint for both hosted and custom)
	endpoint := baseURL + "/payment-gateway/v1.0/debit/payment-host-to-host.htm"
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}
	urlPath := endpointURL.Path

	// Generate signature
	// Format: "<HTTP METHOD>:<RELATIVE PATH URL>:<LOWERCASE_HEX_ENCODED_SHA_256(MINIFIED_HTTP_BODY)>:<X-TIMESTAMP>"
	hash := sha256.New()
	hash.Write(bodyBytes)
	hashedPayload := fmt.Sprintf("%x", hash.Sum(nil))

	stringToSign := fmt.Sprintf("POST:%s:%s:%s", urlPath, hashedPayload, timestamp)

	// Sign with private key
	privateKeyStr := os.Getenv("DANA_PRIVATE_KEY")
	if privateKeyStr == "" {
		return nil, fmt.Errorf("DANA_PRIVATE_KEY is required")
	}

	// Normalize private key (handle \n literals)
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "\\n", "\n")
	if !strings.Contains(privateKeyStr, "-----BEGIN") {
		return nil, fmt.Errorf("invalid private key format: missing PEM headers")
	}

	privateKeyPEM := []byte(privateKeyStr)
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}

	var privateKey *rsa.PrivateKey
	// Try PKCS1 first
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parsed key is not an RSA private key")
		}
		privateKey = rsaKey
	}

	// Create SHA-256 hash of the string to sign
	hash = sha256.New()
	hash.Write([]byte(stringToSign))
	hashed := hash.Sum(nil)

	// Sign the hashed data with PKCS1v15
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode signature as base64
	signature := base64.StdEncoding.EncodeToString(signatureBytes)

	// Get partner ID
	partnerID := getEnv("DANA_X_PARTNER_ID", "")
	if partnerID == "" {
		partnerID = os.Getenv("DANA_CLIENT_ID")
	}

	// Generate external ID
	externalID := "sdk" + uuid.New().String()[3:]

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-SIGNATURE", signature)
	req.Header.Set("X-PARTNER-ID", partnerID)
	req.Header.Set("X-EXTERNAL-ID", externalID)
	req.Header.Set("CHANNEL-ID", "95221")

	// Set optional headers
	if origin := getEnv("DANA_ORIGIN", ""); origin != "" {
		req.Header.Set("ORIGIN", origin)
	}

	// Set debug mode if enabled
	if debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false")); debug {
		if strings.ToLower(env) == "sandbox" {
			req.Header.Set("X-Debug-Mode", "true")
		}
	}

	// Debug: Print request details
	if debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false")); debug {
		fmt.Printf("DEBUG: Raw HTTP Request (Hosted Checkout):\n")
		fmt.Printf("  URL: %s\n", endpoint)
		fmt.Printf("  Method: POST\n")
		fmt.Printf("  Headers:\n")
		for k, v := range req.Header {
			if k == "X-SIGNATURE" {
				fmt.Printf("    %s: %s (truncated)\n", k, v[0][:20]+"...")
			} else {
				fmt.Printf("    %s: %s\n", k, v[0])
			}
		}
		fmt.Printf("  Body:\n%s\n", bodyString)
		fmt.Printf("  String to Sign: %s\n", stringToSign)
	}

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug: Print response
	if debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false")); debug {
		fmt.Printf("DEBUG: Raw HTTP Response (Hosted Checkout):\n")
		fmt.Printf("  Status: %s\n", resp.Status)
		fmt.Printf("  StatusCode: %d\n", resp.StatusCode)
		fmt.Printf("  Body:\n%s\n", string(respBody))
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errData interface{}
		if json.Unmarshal(respBody, &errData) == nil {
			errJSON, _ := json.MarshalIndent(errData, "", "  ")
			return nil, fmt.Errorf("DANA API error (HTTP %d): %s", resp.StatusCode, string(errJSON))
		}
		return nil, fmt.Errorf("DANA API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var response CreateOrderResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
