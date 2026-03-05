package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL         string
	JWTSecret           string
	StripeSecretKey     string
	StripeWebhookSecret string
	R2AccountID         string
	R2AccessKeyID       string
	R2SecretAccessKey   string
	R2BucketName        string
	R2PublicURL         string
	Port                string
	Env                 string
}

func Load() (*Config, error) {
	var missing []string

	require := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	cfg := &Config{
		DatabaseURL:         require("DATABASE_URL"),
		JWTSecret:           require("JWT_SECRET"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		R2AccountID:         os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:       os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:        os.Getenv("R2_BUCKET_NAME"),
		R2PublicURL:         os.Getenv("R2_PUBLIC_URL"),
		Port:                getEnvOr("PORT", "8080"),
		Env:                 getEnvOr("ENV", "development"),
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
