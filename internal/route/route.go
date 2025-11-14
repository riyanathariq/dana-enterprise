package route

import (
	"github.com/gin-gonic/gin"
	"github.com/riyanathariq/dana-enterprise/internal/handler"
)

func SetupRoutes() *gin.Engine {
	// Use gin.New() instead of gin.Default() to avoid duplicate middleware warning
	r := gin.New()

	// Add middleware manually
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Dana Enterprise API is running",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		danaHandler := handler.NewDanaHandler()

		// Merchant routes
		merchant := api.Group("/merchant")
		{
			merchant.GET("/info", danaHandler.GetMerchantInfo)
			merchant.GET("/info/:merchant_id", danaHandler.GetMerchantInfo)
		}

		// Order routes
		order := api.Group("/order")
		{
			order.POST("", danaHandler.CreateOrder)                      // Auto-detect: hosted or custom
			order.POST("/custom", danaHandler.CreateOrderCustomCheckout) // Explicit custom checkout
			// Specific routes must come before parameterized routes
			order.GET("/payment/method", danaHandler.GetPaymentMethod)
			order.GET("/:partner_reference_no", danaHandler.GetOrder)
		}
	}

	return r
}
