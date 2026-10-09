package main

import (
	"fmt"
	"log"
	"os"
)

type config struct {
	DatabaseURL     string
	FrontendURL     string
	Port            string
	GRPCPort        string
	KeycloakJWKSURL string
	MongoURI        string
	MongoDatabase   string
}

func loadConfig() (config, error) {
	if err := LoadDotEnv(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	cfg := config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		FrontendURL:     os.Getenv("FRONTEND_URL"),
		Port:            os.Getenv("PORT"),
		GRPCPort:        os.Getenv("GRPC_PORT"),
		KeycloakJWKSURL: os.Getenv("KC_JWKS_ENDPOINT"),
		MongoURI:        os.Getenv("MONGODB_URI"),
		MongoDatabase:   os.Getenv("MONGODB_DATABASE"),
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
