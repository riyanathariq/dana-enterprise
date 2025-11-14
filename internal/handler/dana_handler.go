package handler

import (
	"net/http"
	"os"

	"github.com/dana-id/dana-go/payment_gateway/v1"
	"github.com/gin-gonic/gin"
	"github.com/riyanathariq/dana-enterprise/internal/mapper"
	"github.com/riyanathariq/dana-enterprise/internal/model"
	"github.com/riyanathariq/dana-enterprise/internal/service/dana/merchant"
	"github.com/riyanathariq/dana-enterprise/internal/service/dana/order"
)

type DanaHandler struct {
	merchantService *merchant.Service
	orderService    *order.Service
}

func NewDanaHandler() *DanaHandler {
	return &DanaHandler{
		merchantService: merchant.NewService(),
		orderService:    order.NewService(),
	}
}

// GetMerchantInfo godoc
// @Summary Get merchant information
// @Description Get merchant resource information including balances
// @Tags merchant
// @Accept json
// @Produce json
// @Param merchant_id path string false "Merchant ID (optional, uses env if not provided)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/merchant/info [get]
// @Router /api/v1/merchant/info/{merchant_id} [get]
func (h *DanaHandler) GetMerchantInfo(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	if merchantID == "" {
		merchantID = os.Getenv("DANA_MERCHANT_ID")
	}

	if merchantID == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "merchant_id is required",
			Code:    "VALIDATION_ERROR",
			Details: "Merchant ID must be provided either as path parameter or in environment variable DANA_MERCHANT_ID",
		})
		return
	}

	result, err := h.merchantService.GetMerchantInfo(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "MERCHANT_INFO_ERROR",
			Details: "Failed to retrieve merchant information from Dana API",
		})
		return
	}

	// Map to clean response format
	response := mapper.MapMerchantResourceResponse(merchantID, result)
	c.JSON(http.StatusOK, response)
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new payment order in DANA Payment Gateway
// @Tags order
// @Accept json
// @Produce json
// @Param request body model.CreateOrderRequest true "Create Order Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/order [post]
func (h *DanaHandler) CreateOrder(c *gin.Context) {
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
			Details: "Invalid request body",
		})
		return
	}

	// Convert request model to service params
	params := order.CreateOrderRequestParams{
		PartnerReferenceNo: req.PartnerReferenceNo,
		MerchantID:         req.MerchantID,
		Amount: payment_gateway.Money{
			Value:    req.Amount.Value,
			Currency: req.Amount.Currency,
		},
		UrlParams:          make([]payment_gateway.UrlParam, len(req.UrlParams)),
		SubMerchantID:      req.SubMerchantID,
		ExternalStoreID:    req.ExternalStoreID,
		ValidUpTo:          req.ValidUpTo,
		DisabledPayMethods: req.DisabledPayMethods,
	}

	// Convert PayOptionDetails (optional for hosted checkout)
	if len(req.PayOptionDetails) > 0 {
		params.PayOptionDetails = make([]payment_gateway.PayOptionDetail, len(req.PayOptionDetails))
		for i, pod := range req.PayOptionDetails {
			payOptionDetail := payment_gateway.PayOptionDetail{
				PayMethod: pod.PayMethod,
				PayOption: pod.PayOption,
				TransAmount: payment_gateway.Money{
					Value:    pod.TransAmount.Value,
					Currency: pod.TransAmount.Currency,
				},
			}
			if pod.FeeAmount != nil {
				payOptionDetail.FeeAmount = &payment_gateway.Money{
					Value:    pod.FeeAmount.Value,
					Currency: pod.FeeAmount.Currency,
				}
			}
			if pod.CardToken != nil {
				payOptionDetail.CardToken = pod.CardToken
			}
			if pod.MerchantToken != nil {
				payOptionDetail.MerchantToken = pod.MerchantToken
			}
			params.PayOptionDetails[i] = payOptionDetail
		}
	}

	// Convert UrlParams
	for i, up := range req.UrlParams {
		params.UrlParams[i] = payment_gateway.UrlParam{
			Url:        up.Url,
			Type:       up.Type,
			IsDeeplink: up.IsDeeplink,
		}
	}

	// Auto-detect: if PayOptionDetails provided, use custom checkout; otherwise use hosted checkout
	var result *payment_gateway.CreateOrderResponse
	var err error

	if len(params.PayOptionDetails) > 0 {
		// Custom Checkout (Host-to-Host)
		result, err = h.orderService.CreateOrderCustomCheckout(c.Request.Context(), params)
	} else {
		// Hosted Checkout (Redirect)
		result, err = h.orderService.CreateOrderHostedCheckout(c.Request.Context(), params)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CREATE_ORDER_ERROR",
			Details: "Failed to create order in Dana API",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order created successfully",
		"data":    result,
	})
}

// CreateOrderCustomCheckout godoc
// @Summary Create a new order using Custom Checkout (Host-to-Host)
// @Description Create a new payment order with specific payment method using DANA Custom Checkout
// @Tags order
// @Accept json
// @Produce json
// @Param request body model.CreateOrderRequest true "Create Order Request (requires pay_option_details)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/order/custom [post]
func (h *DanaHandler) CreateOrderCustomCheckout(c *gin.Context) {
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
			Details: "Invalid request body",
		})
		return
	}

	// Validate PayOptionDetails is required for custom checkout
	if len(req.PayOptionDetails) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "pay_option_details is required for custom checkout",
			Code:    "VALIDATION_ERROR",
			Details: "Custom checkout requires pay_option_details to specify payment method",
		})
		return
	}

	// Convert request model to service params
	params := order.CreateOrderRequestParams{
		PartnerReferenceNo: req.PartnerReferenceNo,
		MerchantID:         req.MerchantID,
		Amount: payment_gateway.Money{
			Value:    req.Amount.Value,
			Currency: req.Amount.Currency,
		},
		PayOptionDetails:   make([]payment_gateway.PayOptionDetail, len(req.PayOptionDetails)),
		UrlParams:          make([]payment_gateway.UrlParam, len(req.UrlParams)),
		SubMerchantID:      req.SubMerchantID,
		ExternalStoreID:    req.ExternalStoreID,
		ValidUpTo:          req.ValidUpTo,
		DisabledPayMethods: req.DisabledPayMethods,
	}

	// Convert PayOptionDetails
	for i, pod := range req.PayOptionDetails {
		payOptionDetail := payment_gateway.PayOptionDetail{
			PayMethod: pod.PayMethod,
			PayOption: pod.PayOption,
			TransAmount: payment_gateway.Money{
				Value:    pod.TransAmount.Value,
				Currency: pod.TransAmount.Currency,
			},
		}
		if pod.FeeAmount != nil {
			payOptionDetail.FeeAmount = &payment_gateway.Money{
				Value:    pod.FeeAmount.Value,
				Currency: pod.FeeAmount.Currency,
			}
		}
		if pod.CardToken != nil {
			payOptionDetail.CardToken = pod.CardToken
		}
		if pod.MerchantToken != nil {
			payOptionDetail.MerchantToken = pod.MerchantToken
		}
		params.PayOptionDetails[i] = payOptionDetail
	}

	// Convert UrlParams
	for i, up := range req.UrlParams {
		params.UrlParams[i] = payment_gateway.UrlParam{
			Url:        up.Url,
			Type:       up.Type,
			IsDeeplink: up.IsDeeplink,
		}
	}

	// Create order using custom checkout
	result, err := h.orderService.CreateOrderCustomCheckout(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CREATE_ORDER_ERROR",
			Details: "Failed to create order in Dana API",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order created successfully (Custom Checkout)",
		"data":    result,
	})
}

// GetPaymentMethod godoc
// @Summary Get payment method
// @Description Get payment method from DANA Payment Gateway
// @Tags payment
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payment/method [get]
func (h *DanaHandler) GetPaymentMethod(c *gin.Context) {
	result, err := h.orderService.GetPaymentMethod(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "GET_PAYMENT_METHOD_ERROR",
			Details: "Failed to get payment method from Dana API",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment method retrieved successfully",
		"data":    result,
	})
}

// GetOrder godoc
// @Summary Get order/payment details
// @Description Query payment order details by partner reference number
// @Tags order
// @Accept json
// @Produce json
// @Param partner_reference_no path string true "Partner Reference Number"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/order/{partner_reference_no} [get]
func (h *DanaHandler) GetOrder(c *gin.Context) {
	partnerReferenceNo := c.Param("partner_reference_no")
	if partnerReferenceNo == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "partner_reference_no is required",
			Code:    "VALIDATION_ERROR",
			Details: "Partner reference number must be provided as path parameter",
		})
		return
	}

	result, err := h.orderService.GetOrder(c.Request.Context(), partnerReferenceNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "GET_ORDER_ERROR",
			Details: "Failed to get order from Dana API",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order retrieved successfully",
		"data":    result,
	})
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *DanaHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Dana Enterprise API is running",
	})
}
