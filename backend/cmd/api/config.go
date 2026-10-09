package main

import (
	"fmt"
	"log"
	"os"
)

type config struct {
	DatabaseURL                 string
	FrontendURL                 string
	Port                        string
	GRPCPort                    string
	KeycloakJWKSURL             string
	KeycloakBaseURL             string
	KeycloakRealm               string
	KeycloakBackendClientID     string
	KeycloakBackendClientSecret string
	MongoURI                    string
	MongoDatabase               string
}

func loadConfig() (config, error) {
	if err := LoadDotEnv(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	cfg := config{
		DatabaseURL:                 os.Getenv("DATABASE_URL"),
		FrontendURL:                 os.Getenv("FRONTEND_URL"),
		Port:                        os.Getenv("PORT"),
		GRPCPort:                    os.Getenv("GRPC_PORT"),
		KeycloakJWKSURL:             os.Getenv("KC_JWKS_ENDPOINT"),
		KeycloakBaseURL:             os.Getenv("KC_BASE_URL"),
		KeycloakRealm:               os.Getenv("KC_REALM"),
		KeycloakBackendClientID:     os.Getenv("KC_BACKEND_CLIENT_ID"),
		KeycloakBackendClientSecret: os.Getenv("KC_BACKEND_CLIENT_SECRET"),
		MongoURI:                    os.Getenv("MONGODB_URI"),
		MongoDatabase:               os.Getenv("MONGODB_DATABASE"),
	}

	if cfg.DatabaseURL == "" {
		return config{}, fmt.Errorf("DATABASE_URL is not set")
	}
	if cfg.KeycloakJWKSURL == "" {
		return config{}, fmt.Errorf("KC_JWKS_ENDPOINT is not set")
	}
	for name, value := range map[string]string{
		"KC_BASE_URL":              cfg.KeycloakBaseURL,
		"KC_REALM":                 cfg.KeycloakRealm,
		"KC_BACKEND_CLIENT_ID":     cfg.KeycloakBackendClientID,
		"KC_BACKEND_CLIENT_SECRET": cfg.KeycloakBackendClientSecret,
	} {
		if value == "" {
			return config{}, fmt.Errorf("%s is not set", name)
		}
	}
	if cfg.MongoURI == "" {
		return config{}, fmt.Errorf("MONGODB_URI is not set")
	}
	if cfg.FrontendURL == "" {
		cfg.FrontendURL = "http://localhost:3000"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.GRPCPort == "" {
		cfg.GRPCPort = "9000"
	}
	if cfg.MongoDatabase == "" {
		cfg.MongoDatabase = "chura"
	}

	return cfg, nil
}
