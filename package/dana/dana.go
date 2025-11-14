package dana

import (
	"os"
	"strconv"
	"sync"

	"github.com/dana-id/dana-go"
	"github.com/dana-id/dana-go/config"
)

var (
	once     sync.Once
	instance *dana.APIClient
)

// getEnv returns environment variable value or default if empty
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// InitData initializes and returns a singleton instance of Dana API Client
func InitData() *dana.APIClient {
	once.Do(func() {
		instance = initializeClient()
	})
	return instance
}

// initializeClient creates a new Dana API client instance
func initializeClient() *dana.APIClient {
	// Parse debug flag from environment
	debug, _ := strconv.ParseBool(getEnv("DANA_DEBUG", "false"))

	// Validate required credentials
	clientID := os.Getenv("DANA_CLIENT_ID")
	privateKey := os.Getenv("DANA_PRIVATE_KEY")
	clientSecret := getEnv("DANA_CLIENT_SECRET", "")

	if clientID == "" {
		panic("DANA_CLIENT_ID is required but not set in environment variables")
	}
	if privateKey == "" {
		panic("DANA_PRIVATE_KEY is required but not set in environment variables")
	}
	if clientSecret == "" {
		panic("DANA_CLIENT_SECRET is required but not set in environment variables")
	}

	// X_PARTNER_ID should be Client ID for authentication, not Merchant ID
	// Try in order: explicit X_PARTNER_ID -> Client ID -> Merchant ID
	partnerID := getEnv("DANA_X_PARTNER_ID", "")
	if partnerID == "" {
		// Use Client ID as Partner ID (most common case for authentication)
		partnerID = clientID
	}

	// Determine server URL based on environment
	env := getEnv("DANA_ENV", "sandbox")
	var serverURL string
	if env == "production" {
		serverURL = "https://api.dana.id"
	} else {
		// Use custom host if provided, otherwise use default sandbox URL
		if host := getEnv("DANA_HOST", ""); host != "" {
			scheme := getEnv("DANA_SCHEME", "https")
			serverURL = scheme + "://" + host
		} else {
			serverURL = "https://api.sandbox.dana.id"
		}
	}

	return dana.NewAPIClient(&config.Configuration{
		Host:          getEnv("DANA_HOST", ""),
		Scheme:        getEnv("DANA_SCHEME", "https"),
		DefaultHeader: nil,
		UserAgent:     getEnv("DANA_USER_AGENT", ""),
		Debug:         debug,
		Servers: config.ServerConfigurations{
			{
				URL:         serverURL,
				Description: "DANA API Gateway " + env,
			},
		},
		OperationServers: nil,
		HTTPClient:       nil,
		APIKey: &config.APIKey{
			ENV:              env,
			DANA_ENV:         env,
			ORIGIN:           getEnv("DANA_ORIGIN", ""),
			X_PARTNER_ID:     partnerID,
			CHANNEL_ID:       getEnv("DANA_CHANNEL_ID", ""),
			PRIVATE_KEY:      os.Getenv("DANA_PRIVATE_KEY"),
			PRIVATE_KEY_PATH: getEnv("DANA_PRIVATE_KEY_PATH", ""),
			CLIENT_SECRET:    getEnv("DANA_CLIENT_SECRET", ""),
			CLIENT_ID:        os.Getenv("DANA_CLIENT_ID"),
			ACCESS_TOKEN:     getEnv("DANA_ACCESS_TOKEN", ""),
			X_DEBUG:          getEnv("DANA_X_DEBUG", ""),
		},
	})
}
