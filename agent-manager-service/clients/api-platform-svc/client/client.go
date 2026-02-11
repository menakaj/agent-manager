//
// Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//

// Package client provides the Platform API client for gateway management.
//
//go:generate moq -rm -fmt goimports -skip-ensure -pkg clientmocks -out ../../clientmocks/platformapi_client_fake.go . PlatformAPIClient:PlatformAPIClientMock
package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/wso2/ai-agent-management-platform/agent-manager-service/clients/api-platform-svc/gen"
	"github.com/wso2/ai-agent-management-platform/agent-manager-service/clients/requests"
)

// Config contains configuration for the Platform API client
type Config struct {
	BaseURL      string
	AuthProvider AuthProvider
	RetryConfig  requests.RequestRetryConfig
}

// PlatformAPIClient defines the interface for Platform API operations
type PlatformAPIClient interface {
	// Gateway Operations
	CreateGateway(ctx context.Context, req CreateGatewayRequest) (*GatewayResponse, error)
	GetGateway(ctx context.Context, gatewayID string) (*GatewayResponse, error)
	ListGateways(ctx context.Context, filters GatewayFilters) (*GatewayListResponse, error)
	UpdateGateway(ctx context.Context, gatewayID string, req UpdateGatewayRequest) (*GatewayResponse, error)
	DeleteGateway(ctx context.Context, gatewayID string) error
	RotateGatewayToken(ctx context.Context, gatewayID string) (*TokenRotationResponse, error)
	RevokeGatewayToken(ctx context.Context, gatewayID, tokenID string) error

	// LLM Provider Operations
	CreateLLMProvider(ctx context.Context, req CreateLLMProviderRequest) (*LLMProviderResponse, error)
	GetLLMProvider(ctx context.Context, providerID string) (*LLMProviderResponse, error)
	ListLLMProviders(ctx context.Context, limit, offset int32) (*LLMProviderListResponse, error)
	UpdateLLMProvider(ctx context.Context, providerID string, req UpdateLLMProviderRequest) (*LLMProviderResponse, error)
	DeleteLLMProvider(ctx context.Context, providerID string) error

	// LLM Provider Template Operations
	GetLLMProviderTemplate(ctx context.Context, templateID string) (*LLMProviderTemplateResponse, error)
	ListLLMProviderTemplates(ctx context.Context, limit, offset int32) (*LLMProviderTemplateListResponse, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

type platformAPIClient struct {
	baseURL   string
	apiClient *gen.ClientWithResponses
}

// NewPlatformAPIClient creates a new Platform API client
func NewPlatformAPIClient(cfg *Config) (PlatformAPIClient, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.AuthProvider == nil {
		return nil, fmt.Errorf("auth provider is required")
	}

	// Create the retryable HTTP client (uses defaults if RetryConfig is zero-value)
	httpClient := requests.NewRetryableHTTPClient(&http.Client{}, cfg.RetryConfig)

	// Create auth request editor
	authEditor := func(ctx context.Context, req *http.Request) error {
		slog.Debug("Adding auth token to Platform API request")
		token, err := cfg.AuthProvider.GetToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	// Create the generated OpenAPI client with retryable HTTP client and auth
	apiClient, err := gen.NewClientWithResponses(
		cfg.BaseURL,
		gen.WithHTTPClient(httpClient),
		gen.WithRequestEditorFn(authEditor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Platform API client: %w", err)
	}

	return &platformAPIClient{
		baseURL:   cfg.BaseURL,
		apiClient: apiClient,
	}, nil
}

// HealthCheck performs a health check on the Platform API
func (c *platformAPIClient) HealthCheck(ctx context.Context) error {
	// TODO: Implement health check endpoint when available in Platform API
	// For now, try to list gateways as a simple health check
	resp, err := c.apiClient.ListGatewaysWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.StatusCode() >= 500 {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode())
	}

	return nil
}
