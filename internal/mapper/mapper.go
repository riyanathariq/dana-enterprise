package mapper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dana-id/dana-go/merchant_management/v1"
	"github.com/riyanathariq/dana-enterprise/internal/model"
)

// MapMerchantResourceResponse maps Dana API response to our clean response model
func MapMerchantResourceResponse(merchantID string, danaResponse *merchant_management.QueryMerchantResourceResponse) *model.MerchantInfoResponse {
	if danaResponse == nil {
		return &model.MerchantInfoResponse{
			Success: false,
			Message: "No data available",
		}
	}

	response := &model.MerchantInfoResponse{
		Success: true,
		Message: "Merchant information retrieved successfully",
		Data: &model.MerchantInfoData{
			MerchantID: merchantID,
			Balances:   &model.MerchantBalances{},
			Resources:  make(map[string]*model.MerchantResource),
		},
		Meta: &model.MerchantInfoMeta{
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Extract response
	resp := danaResponse.Response

	// Extract request ID from head if available
	head := resp.Head
	if head.ReqMsgId != nil && *head.ReqMsgId != "" {
		response.Meta.RequestID = *head.ReqMsgId
	}

	// Map merchant resource information
	body := resp.Body
	if len(body.MerchantResourceInformations) > 0 {
		for _, info := range body.MerchantResourceInformations {
			resourceType := info.GetResourceType()
			resourceValue := info.GetValue()

			if resourceType != "" && resourceValue != "" {
				// Parse value JSON to get amount and currency
				var valueData map[string]interface{}
				if err := json.Unmarshal([]byte(resourceValue), &valueData); err == nil {
					// Extract amount
					amount := ""
					if amountVal, ok := valueData["amount"]; ok {
						if amountStr, ok := amountVal.(string); ok {
							amount = amountStr
						} else if amountFloat, ok := amountVal.(float64); ok {
							amount = formatAmount(amountFloat)
						}
					}

					currency := "IDR"
					if currencyVal, ok := valueData["currency"]; ok {
						if currencyStr, ok := currencyVal.(string); ok {
							currency = currencyStr
						}
					}

					balanceInfo := &model.BalanceInfo{
						Amount:   amount,
						Currency: currency,
					}

					// Map to balances based on type
					switch resourceType {
					case "MERCHANT_DEPOSIT_BALANCE":
						response.Data.Balances.DepositBalance = balanceInfo
					case "MERCHANT_AVAILABLE_BALANCE":
						response.Data.Balances.AvailableBalance = balanceInfo
					case "MERCHANT_TOTAL_BALANCE":
						response.Data.Balances.TotalBalance = balanceInfo
					}

					// Store in resources map
					response.Data.Resources[resourceType] = &model.MerchantResource{
						Type:        resourceType,
						Value:       amount,
						Description: getResourceDescription(resourceType),
					}
				} else {
					// If value is not JSON, use as is
					response.Data.Resources[resourceType] = &model.MerchantResource{
						Type:        resourceType,
						Value:       resourceValue,
						Description: getResourceDescription(resourceType),
					}
				}
			}
		}
	}

	// If no resources found, indicate that
	if len(response.Data.Resources) == 0 {
		response.Message = "No merchant resource information found"
		response.Data.Resources["note"] = &model.MerchantResource{
			Type:        "info",
			Value:       "No resources available",
			Description: "The merchant may not have any resource information available at this time",
		}
	}

	return response
}

// getResourceDescription returns a human-readable description for resource types
func getResourceDescription(resourceType string) string {
	descriptions := map[string]string{
		"MERCHANT_DEPOSIT_BALANCE":   "Total deposit balance of the merchant",
		"MERCHANT_AVAILABLE_BALANCE": "Available balance that can be used for transactions",
		"MERCHANT_TOTAL_BALANCE":     "Total balance including all account balances",
	}

	if desc, ok := descriptions[resourceType]; ok {
		return desc
	}
	return ""
}

// formatAmount formats float64 amount to string
func formatAmount(amount float64) string {
	// Format to string without unnecessary decimal places
	return fmt.Sprintf("%.0f", amount)
}
