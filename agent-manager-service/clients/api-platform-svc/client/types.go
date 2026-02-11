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

package client

import "time"

// -----------------------------------------------------------------------------
// Gateway Request/Response Types
// -----------------------------------------------------------------------------

// CreateGatewayRequest contains data for registering a new gateway
type CreateGatewayRequest struct {
	Name              string
	DisplayName       string
	Description       string
	VHost             string
	FunctionalityType string // "ai", "regular", "event"
	IsCritical        bool
	Properties        map[string]interface{}
}

// UpdateGatewayRequest contains data for updating a gateway
type UpdateGatewayRequest struct {
	DisplayName string
	Description string
	IsCritical  *bool
	Properties  map[string]interface{}
}

// GatewayResponse represents a gateway from Platform API
type GatewayResponse struct {
	ID                string
	OrganizationID    string
	Name              string
	DisplayName       string
	Description       string
	VHost             string
	FunctionalityType string
	IsCritical        bool
	IsActive          bool
	Properties        map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GatewayListResponse represents a list of gateways
type GatewayListResponse struct {
	Gateways   []GatewayResponse
	Total      int32
	Limit      int32
	Offset     int32
	Pagination PaginationInfo
}

// GatewayFilters for querying gateways
type GatewayFilters struct {
	FunctionalityType *string
	IsActive          *bool
	Limit             *int32
	Offset            *int32
}

// TokenRotationResponse represents response from token rotation
type TokenRotationResponse struct {
	TokenID   string
	Token     string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

// -----------------------------------------------------------------------------
// LLM Provider Request/Response Types
// -----------------------------------------------------------------------------

// CreateLLMProviderRequest contains data for creating an LLM provider
type CreateLLMProviderRequest struct {
	ID             string
	Name           string
	DisplayName    string
	Description    string
	Version        string
	Template       string
	Upstream       map[string]interface{}
	Context        string
	VHost          string
	AccessControl  *LLMAccessControl
	OpenAPI        string
	ModelProviders []LLMModelProvider
	RateLimiting   *LLMRateLimiting
}

// UpdateLLMProviderRequest contains data for updating an LLM provider
type UpdateLLMProviderRequest struct {
	DisplayName    *string
	Description    *string
	Version        *string
	Upstream       map[string]interface{}
	AccessControl  *LLMAccessControl
	OpenAPI        *string
	ModelProviders *[]LLMModelProvider
	RateLimiting   *LLMRateLimiting
}

// LLMProviderResponse represents an LLM provider from Platform API
type LLMProviderResponse struct {
	ID             string
	OrganizationID string
	Name           string
	DisplayName    string
	Description    string
	Version        string
	Template       string
	Upstream       map[string]interface{}
	Context        string
	VHost          string
	AccessControl  *LLMAccessControl
	OpenAPI        string
	ModelProviders []LLMModelProvider
	RateLimiting   *LLMRateLimiting
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// LLMProviderListResponse represents a list of LLM providers
type LLMProviderListResponse struct {
	Providers  []LLMProviderResponse
	Total      int32
	Limit      int32
	Offset     int32
	Pagination PaginationInfo
}

// LLMAccessControl defines access control for LLM provider
type LLMAccessControl struct {
	Mode       string // "allow_all" or "deny_all"
	Exceptions []LLMAccessControlRule
}

// LLMAccessControlRule defines an access control rule
type LLMAccessControlRule struct {
	Path    string
	Methods []string
}

// LLMModelProvider defines a model provider configuration
type LLMModelProvider struct {
	ID     string
	Name   string
	Models []LLMModel
}

// LLMModel defines an LLM model
type LLMModel struct {
	ID          string
	Name        string
	Description string
}

// LLMRateLimiting defines rate limiting configuration
type LLMRateLimiting struct {
	ProviderLevel *LLMProviderLevelLimits
}

// LLMProviderLevelLimits defines provider-level rate limits
type LLMProviderLevelLimits struct {
	Global *LLMGlobalLimits
}

// LLMGlobalLimits defines global rate limits
type LLMGlobalLimits struct {
	Request *LLMRateLimit
	Token   *LLMRateLimit
}

// LLMRateLimit defines a rate limit
type LLMRateLimit struct {
	Enabled bool
	Count   int64
	Reset   *LLMResetDuration
}

// LLMResetDuration defines reset duration for rate limit
type LLMResetDuration struct {
	Duration int64
	Unit     string // "second", "minute", "hour", "day", "week", "month"
}

// -----------------------------------------------------------------------------
// LLM Provider Template Types
// -----------------------------------------------------------------------------

// LLMProviderTemplateResponse represents an LLM provider template
type LLMProviderTemplateResponse struct {
	ID            string
	Name          string
	DisplayName   string
	Description   string
	Version       string
	BaseURL       string
	DefaultModels []LLMModel
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LLMProviderTemplateListResponse represents a list of templates
type LLMProviderTemplateListResponse struct {
	Templates  []LLMProviderTemplateResponse
	Total      int32
	Limit      int32
	Offset     int32
	Pagination PaginationInfo
}

// -----------------------------------------------------------------------------
// Common Types
// -----------------------------------------------------------------------------

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Limit  int32
	Offset int32
	Total  int32
}
