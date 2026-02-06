package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	APIKey string

	WhatsAppAPIURL        string
	WhatsAppPhoneNumberID string
	WhatsAppAccessToken   string
	WhatsAppTemplateName  string
	WhatsAppTemplateLang  string
	WhatsAppRecipient     string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, reading from environment")
	}

	cfg := &Config{
		AppPort:               getEnvOrDefault("APP_PORT", ":8080"),
		AppEnv:                getEnvOrDefault("APP_ENV", "development"),
		APIKey:                os.Getenv("API_KEY"),
		WhatsAppAPIURL:        getEnvOrDefault("WHATSAPP_API_URL", "https://graph.facebook.com/v21.0"),
		WhatsAppPhoneNumberID: os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		WhatsAppAccessToken:   os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		WhatsAppTemplateName:  getEnvOrDefault("WHATSAPP_TEMPLATE_NAME", "order_notification"),
		WhatsAppTemplateLang:  getEnvOrDefault("WHATSAPP_TEMPLATE_LANGUAGE", "en"),
		WhatsAppRecipient:     os.Getenv("WHATSAPP_RECIPIENT_NUMBER"),
	}

	if err := cfg.validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	return cfg
}

func (c *Config) validate() error {
	if c.WhatsAppPhoneNumberID == "" {
		return fmt.Errorf("WHATSAPP_PHONE_NUMBER_ID is required")
	}
	if c.WhatsAppAccessToken == "" {
		return fmt.Errorf("WHATSAPP_ACCESS_TOKEN is required")
	}
	if c.WhatsAppRecipient == "" {
		return fmt.Errorf("WHATSAPP_RECIPIENT_NUMBER is required")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
