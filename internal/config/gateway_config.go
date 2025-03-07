package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type GatewayEndpoints struct {
	Deposit    string `yaml:"deposit"`
	Withdrawal string `yaml:"withdrawal"`
}

type GatewayRetry struct {
	MaxAttempts   int     `yaml:"max_attempts"`
	BackoffFactor float64 `yaml:"backoff_factor"`
}

type GatewayDetails struct {
	BaseURL     string            `yaml:"base_url"`
	Endpoints   GatewayEndpoints  `yaml:"endpoints"`
	CallbackURL string            `yaml:"callback_url"`
	Headers     map[string]string `yaml:"headers"`
	Timeout     int               `yaml:"timeout"`
	Retry       GatewayRetry      `yaml:"retry"`
}

type CountryConfig struct {
	Gateways map[string]int `yaml:"gateways"`
}

type GatewayConfig struct {
	Gateways  map[string]GatewayDetails `yaml:"gateways"`
	Countries map[string]CountryConfig  `yaml:"countries"`
}

// GetGatewayDetails returns the gateway details for a given gateway name
func (c *GatewayConfig) GetGatewayDetails(gatewayName string) (GatewayDetails, bool) {
	details, exists := c.Gateways[gatewayName]
	return details, exists
}

type GatewayPriority struct {
	Name     string
	ID       int
	Priority int
}

func LoadGatewayConfig(configPath string) (*GatewayConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config GatewayConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *GatewayConfig) error {
	if len(config.Gateways) == 0 {
		return fmt.Errorf("no gateways defined in configuration")
	}

	if len(config.Countries) == 0 {
		return fmt.Errorf("no countries defined in configuration")
	}

	// Validate that all gateways referenced in countries exist
	for countryCode, country := range config.Countries {

		if len(country.Gateways) == 0 {
			return fmt.Errorf("no gateways defined for country %s", countryCode)
		}

		for gatewayName := range country.Gateways {
			if _, exists := config.Gateways[gatewayName]; !exists {
				return fmt.Errorf("gateway %s referenced in country %s does not exist in gateways configuration",
					gatewayName, countryCode)
			}
		}
	}

	return nil
}
