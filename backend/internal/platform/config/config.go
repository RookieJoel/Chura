package config

import (
	"fmt"
	"log"
	"os"
)

// Service identifies which domain binary is running.
type Service string

const (
	ServiceProject   Service = "project"
	ServiceSprint    Service = "sprint"
	ServiceAnalytics Service = "analytics"
	ServiceWorkitem  Service = "workitem"
)

type Config struct {
	Service Service

	DatabaseURL                 string
	FrontendURL                 string
	Port                        string
	KeycloakJWKSURL             string
	KeycloakBaseURL             string
	KeycloakRealm               string
	KeycloakBackendClientID     string
	KeycloakBackendClientSecret string
	MongoURI                    string
	MongoDatabase               string
	SMTPHost                    string
	SMTPPort                    string
	SMTPUsername                string
	SMTPPassword                string
	SMTPFrom                    string
	RabbitMQURL                 string
	RabbitMQNotificationQueue   string
	InstanceName                string

	ConsulHTTPAddr      string
	ConsulServiceName   string
	ConsulServicePort   int
	ConsulServiceAddress string
}

func defaultPort(service Service) string {
	switch service {
	case ServiceAnalytics:
		return "8056"
	case ServiceSprint:
		return "8057"
	case ServiceWorkitem:
		return "8058"
	case ServiceProject:
		return "8059"
	default:
		return "8080"
	}
}

func defaultConsulServiceName(service Service) string {
	switch service {
	case ServiceAnalytics:
		return "chura-analytics"
	case ServiceSprint:
		return "chura-sprint"
	case ServiceWorkitem:
		return "chura-workitem"
	case ServiceProject:
		return "chura-project"
	default:
		return ""
	}
}

func Load(service Service) (Config, error) {
	envFile := os.Getenv("CHURA_ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	if err := LoadDotEnv(envFile); err != nil {
		log.Printf("no environment file %q loaded: %v", envFile, err)
	}

	cfg := Config{
		Service:                     service,
		DatabaseURL:                 os.Getenv("DATABASE_URL"),
		FrontendURL:                 os.Getenv("FRONTEND_URL"),
		Port:                        os.Getenv("PORT"),
		KeycloakJWKSURL:             os.Getenv("KC_JWKS_ENDPOINT"),
		KeycloakBaseURL:             os.Getenv("KC_BASE_URL"),
		KeycloakRealm:               os.Getenv("KC_REALM"),
		KeycloakBackendClientID:     os.Getenv("KC_BACKEND_CLIENT_ID"),
		KeycloakBackendClientSecret: os.Getenv("KC_BACKEND_CLIENT_SECRET"),
		MongoURI:                    os.Getenv("MONGODB_URI"),
		MongoDatabase:               os.Getenv("MONGODB_DATABASE"),
		SMTPHost:                    os.Getenv("SMTP_HOST"),
		SMTPPort:                    os.Getenv("SMTP_PORT"),
		SMTPUsername:                os.Getenv("SMTP_USERNAME"),
		SMTPPassword:                os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                    os.Getenv("SMTP_FROM"),
		RabbitMQURL:                 os.Getenv("RABBITMQ_URL"),
		RabbitMQNotificationQueue:   os.Getenv("RABBITMQ_NOTIFICATION_QUEUE"),
		InstanceName:                os.Getenv("INSTANCE_NAME"),
		ConsulHTTPAddr:              os.Getenv("CONSUL_HTTP_ADDR"),
		ConsulServiceName:           os.Getenv("CONSUL_SERVICE_NAME"),
		ConsulServiceAddress:        os.Getenv("CONSUL_SERVICE_ADDRESS"),
	}

	if cfg.FrontendURL == "" {
		cfg.FrontendURL = "http://localhost:3000"
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort(service)
	}
	if cfg.MongoDatabase == "" {
		cfg.MongoDatabase = "chura"
	}
	if cfg.RabbitMQNotificationQueue == "" {
		cfg.RabbitMQNotificationQueue = "chura-notifications"
	}
	if cfg.InstanceName == "" {
		cfg.InstanceName, _ = os.Hostname()
	}
	if cfg.ConsulServiceName == "" {
		cfg.ConsulServiceName = defaultConsulServiceName(service)
	}
	if port := os.Getenv("CONSUL_SERVICE_PORT"); port != "" {
		if _, err := fmt.Sscanf(port, "%d", &cfg.ConsulServicePort); err != nil {
			return Config{}, fmt.Errorf("CONSUL_SERVICE_PORT: %w", err)
		}
	} else if p, err := parsePort(cfg.Port); err == nil {
		cfg.ConsulServicePort = p
	}

	if err := cfg.validate(service); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func parsePort(port string) (int, error) {
	var p int
	if _, err := fmt.Sscanf(port, "%d", &p); err != nil {
		return 0, err
	}
	return p, nil
}

func (cfg Config) validate(service Service) error {
	if cfg.KeycloakJWKSURL == "" {
		return fmt.Errorf("KC_JWKS_ENDPOINT is not set")
	}
	for name, value := range map[string]string{
		"KC_BASE_URL":              cfg.KeycloakBaseURL,
		"KC_REALM":                 cfg.KeycloakRealm,
		"KC_BACKEND_CLIENT_ID":     cfg.KeycloakBackendClientID,
		"KC_BACKEND_CLIENT_SECRET": cfg.KeycloakBackendClientSecret,
	} {
		if value == "" {
			return fmt.Errorf("%s is not set", name)
		}
	}

	switch service {
	case ServiceProject, ServiceSprint:
		if cfg.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is not set")
		}
	case ServiceAnalytics:
		if cfg.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is not set")
		}
		if cfg.MongoURI == "" {
			return fmt.Errorf("MONGODB_URI is not set")
		}
	case ServiceWorkitem:
		if cfg.MongoURI == "" {
			return fmt.Errorf("MONGODB_URI is not set")
		}
		if cfg.RabbitMQURL == "" {
			return fmt.Errorf("RABBITMQ_URL is not set")
		}
	}

	return nil
}
