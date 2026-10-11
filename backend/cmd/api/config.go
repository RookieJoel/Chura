package main

import (
	"fmt"
	"log"
	"os"
)

type config struct {
	DatabaseURL               string
	FrontendURL               string
	Port                      string
	KeycloakJWKSURL           string
	MongoURI                  string
	MongoDatabase             string
	TestEmail                 string
	SMTPHost                  string
	SMTPPort                  string
	SMTPUsername              string
	SMTPPassword              string
	SMTPFrom                  string
	RabbitMQURL               string
	RabbitMQNotificationQueue string
}

func loadConfig() (config, error) {
	if err := LoadDotEnv(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	cfg := config{
		DatabaseURL:               os.Getenv("DATABASE_URL"),
		FrontendURL:               os.Getenv("FRONTEND_URL"),
		Port:                      os.Getenv("PORT"),
		KeycloakJWKSURL:           os.Getenv("KC_JWKS_ENDPOINT"),
		MongoURI:                  os.Getenv("MONGODB_URI"),
		MongoDatabase:             os.Getenv("MONGODB_DATABASE"),
		TestEmail:                 os.Getenv("TEST_EMAIL_TO"),
		SMTPHost:                  os.Getenv("SMTP_HOST"),
		SMTPPort:                  os.Getenv("SMTP_PORT"),
		SMTPUsername:              os.Getenv("SMTP_USERNAME"),
		SMTPPassword:              os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                  os.Getenv("SMTP_FROM"),
		RabbitMQURL:               os.Getenv("RABBITMQ_URL"),
		RabbitMQNotificationQueue: os.Getenv("RABBITMQ_NOTIFICATION_QUEUE"),
	}

	if cfg.DatabaseURL == "" {
		return config{}, fmt.Errorf("DATABASE_URL is not set")
	}
	if cfg.KeycloakJWKSURL == "" {
		return config{}, fmt.Errorf("KC_JWKS_ENDPOINT is not set")
	}
	if cfg.MongoURI == "" {
		return config{}, fmt.Errorf("MONGODB_URI is not set")
	}
	if cfg.RabbitMQURL == "" {
		return config{}, fmt.Errorf("RABBITMQ_URL is not set")
	}
	if cfg.FrontendURL == "" {
		cfg.FrontendURL = "http://localhost:3000"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.MongoDatabase == "" {
		cfg.MongoDatabase = "chura"
	}
	if cfg.RabbitMQNotificationQueue == "" {
		cfg.RabbitMQNotificationQueue = "chura-notifications"
	}

	return cfg, nil
}
