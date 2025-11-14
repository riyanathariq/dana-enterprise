package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/riyanathariq/dana-enterprise/internal/route"
	"github.com/riyanathariq/dana-enterprise/package/dana"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: .env file not found: %v\n", err)
	}

	// Initialize Dana API Client
	danaClient := dana.InitData()
	if danaClient == nil {
		log.Fatal("❌ Failed to initialize Dana API client")
	}

	// Print configuration
	fmt.Println("✅ Dana API Client initialized successfully!")
	fmt.Println("📋 Configuration:")
	if env := os.Getenv("DANA_ENV"); env != "" {
		fmt.Printf("   - Environment: %s\n", env)
	}
	fmt.Printf("   - Client ID: %s\n", os.Getenv("DANA_CLIENT_ID"))
	if merchantID := os.Getenv("DANA_MERCHANT_ID"); merchantID != "" {
		fmt.Printf("   - Merchant ID: %s\n", merchantID)
	}

	// Set Gin mode (production/release mode)
	if os.Getenv("GIN_MODE") == "" {
		// Default to release mode if not set
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup routes
	r := route.SetupRoutes()

	// Trust only localhost proxies in development
	// In production, set specific trusted proxies
	if os.Getenv("GIN_TRUSTED_PROXIES") != "" {
		r.SetTrustedProxies([]string{os.Getenv("GIN_TRUSTED_PROXIES")})
	} else {
		r.SetTrustedProxies(nil) // Don't trust all proxies
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3150"
	}

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
