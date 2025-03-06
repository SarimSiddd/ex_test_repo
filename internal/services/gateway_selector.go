package services

import (
	"context"
	"fmt"
	"payment-gateway/internal/config"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
)

type GatewaySelector struct {
	gatewayConfig *config.GatewayConfig
	countryRepo   repository.Country
	gatewayRepo   repository.Gateway
	userRepo      repository.User
}

func NewGatewaySelector(
	gatewayConfig *config.GatewayConfig,
	countryRepo repository.Country,
	gatewayRepo repository.Gateway,
) *GatewaySelector {
	return &GatewaySelector{
		gatewayConfig: gatewayConfig,
		countryRepo:   countryRepo,
		gatewayRepo:   gatewayRepo,
	}
}

func (s *GatewaySelector) SelectGateway(ctx context.Context, countryCode string) (*models.Gateway, error) {

	countryConfig, exists := s.gatewayConfig.Countries[countryCode]

	if !exists {
		return nil, fmt.Errorf("no configuration found for country %s", countryCode)
	}

	if len(countryConfig.Gateways) == 0 {
		return nil, fmt.Errorf("no gateways defined for country: %s", countryCode)
	}

	var highestPriority int = -1
	var highestPriorityGateway string

	for gateway, priority := range countryConfig.Gateways {
		if priority > highestPriority {
			highestPriority = priority
			highestPriorityGateway = gateway
		}
	}

	gateway, err := s.gatewayRepo.FindByName(ctx, highestPriorityGateway)

	if err != nil {
		return nil, fmt.Errorf("failed to find gateway: %w", err)
	}

	return gateway, nil
}

func (s *GatewaySelector) SelectGatewayForUser(ctx context.Context, userID int) (*models.Gateway, error) {

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	country, err := s.countryRepo.FindByID(ctx, user.CountryID)
	if err != nil {
		return nil, fmt.Errorf("failed to find country for user: %w", err)
	}

	return s.SelectGateway(ctx, country.Code)
}
