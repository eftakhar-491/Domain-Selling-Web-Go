package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

// EnvConfig holds all environment variables configuration
type EnvConfig struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	EncryptionKey       string
	RedisURL            string
	SMTPHost            string
	SMTPPort            int
	SMTPEmail           string
	SMTPPassword        string
	ResellerUsername    string
	DNAAccountUsername  string
	ResellerPassword    string
	DNATestAPIKey       string
	DNAProdAPIKey       string
	ResellerTestMode    bool
	ResellerBaseURL     string
	DNATestRestBaseURL  string
	DNAProdRestBaseURL  string
	DNAProdSOAPEndpoint string
	DNATestSOAPEndpoint string
}

var (
	envInstance *EnvConfig
	envOnce     sync.Once
)

// LoadEnv reads .env and returns the EnvConfig singleton
func LoadEnv() *EnvConfig {
	envOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("Note: .env file not found or already loaded into environment")
		}

		testMode, _ := strconv.ParseBool(os.Getenv("RESELLER_TEST_MODE"))
		smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
		if smtpPort == 0 {
			smtpPort = 587
		}

		envInstance = &EnvConfig{
			Port:                os.Getenv("PORT"),
			DatabaseURL:         os.Getenv("DATABASE_URL"),
			JWTSecret:           os.Getenv("JWT_SECRET"),
			EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
			RedisURL:            os.Getenv("REDIS_URL"),
			SMTPHost:            os.Getenv("SMTP_HOST"),
			SMTPPort:            smtpPort,
			SMTPEmail:           os.Getenv("SMTP_EMAIL"),
			SMTPPassword:        os.Getenv("SMTP_PASSWORD"),
			ResellerUsername:    os.Getenv("RESELLER_USERNAME"),
			DNAAccountUsername:  os.Getenv("DNA_ACCOUNT_USERNAME"),
			ResellerPassword:    os.Getenv("RESELLER_PASSWORD"),
			DNATestAPIKey:       os.Getenv("DNA_TEST_API_KEY"),
			DNAProdAPIKey:       os.Getenv("DNA_PROD_API_KEY"),
			ResellerTestMode:    testMode,
			ResellerBaseURL:     os.Getenv("RESELLER_BASE_URL"),
			DNATestRestBaseURL:  os.Getenv("DNA_TEST_REST_BASE_URL"),
			DNAProdRestBaseURL:  os.Getenv("DNA_PROD_REST_BASE_URL"),
			DNAProdSOAPEndpoint: os.Getenv("DNA_PROD_SOAP_ENDPOINT"),
			DNATestSOAPEndpoint: os.Getenv("DNA_TEST_SOAP_ENDPOINT"),
		}
	})

	return envInstance
}

// GetEnv returns the current EnvConfig, initializing it if needed
func GetEnv() *EnvConfig {
	if envInstance == nil {
		return LoadEnv()
	}
	return envInstance
}
