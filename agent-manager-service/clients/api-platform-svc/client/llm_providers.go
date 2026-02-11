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

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/ai-agent-management-platform/agent-manager-service/clients/api-platform-svc/gen"
	"github.com/wso2/ai-agent-management-platform/agent-manager-service/utils"
)

// CreateLLMProvider creates a new LLM provider in Platform API
func (c *platformAPIClient) CreateLLMProvider(ctx context.Context, req CreateLLMProviderRequest) (*LLMProviderResponse, error) {
	apiReq := gen.CreateLLMProviderJSONRequestBody{
		Id:       req.ID,
		Name:     req.Name,
		Template: req.Template,
		Version:  req.Version,
	}

	// Context & VHost are optional pointer fields
	if req.Context != "" {
		apiReq.Context = &req.Context
	}
	if req.VHost != "" {
		apiReq.Vhost = &req.VHost
	}

	// Description & OpenAPI are optional pointer fields
	if req.Description != "" {
		apiReq.Description = &req.Description
	}
	if req.OpenAPI != "" {
		apiReq.Openapi = &req.OpenAPI
	}

	// Map upstream (raw map -> typed upstream)
	if req.Upstream != nil {
		upstream, err := mapToGenUpstream(req.Upstream)
		if err != nil {
			return nil, fmt.Errorf("invalid upstream configuration: %w", err)
		}
		apiReq.Upstream = upstream
	}

	// Map access control
	if ac := buildGenLLMAccessControl(req.AccessControl); ac != nil {
		apiReq.AccessControl = *ac
	}

	// Map model providers
	if len(req.ModelProviders) > 0 {
		modelProviders := make([]gen.LLMModelProvider, 0, len(req.ModelProviders))
		for _, mp := range req.ModelProviders {
			genModels := make([]gen.LLMModel, 0, len(mp.Models))
			for _, m := range mp.Models {
				name := m.Name
				model := gen.LLMModel{
					Id: m.ID,
				}
				if name != "" {
					model.Name = &name
				}
				if m.Description != "" {
					desc := m.Description
					model.Description = &desc
				}
				genModels = append(genModels, model)
			}

			provider := gen.LLMModelProvider{
				Id: mp.ID,
			}
			if len(genModels) > 0 {
				provider.Models = &genModels
			}
			if mp.Name != "" {
				name := mp.Name
				provider.Name = &name
			}

			modelProviders = append(modelProviders, provider)
		}
		apiReq.ModelProviders = &modelProviders
	}

	// Map rate limiting
	if rl := buildGenLLMRateLimitingConfig(req.RateLimiting); rl != nil {
		apiReq.RateLimiting = rl
	}

	resp, err := c.apiClient.CreateLLMProviderWithResponse(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			ConflictErr: utils.ErrLLMProviderAlreadyExists,
		})
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("empty response from create LLM provider")
	}

	return toLLMProviderResponse(resp.JSON201), nil
}

// GetLLMProvider retrieves an LLM provider by ID from Platform API
func (c *platformAPIClient) GetLLMProvider(ctx context.Context, providerID string) (*LLMProviderResponse, error) {
	resp, err := c.apiClient.GetLLMProviderWithResponse(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrLLMProviderNotFound,
		})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from get LLM provider")
	}

	return toLLMProviderResponse(resp.JSON200), nil
}

// ListLLMProviders lists LLM providers with pagination
func (c *platformAPIClient) ListLLMProviders(ctx context.Context, limit, offset int32) (*LLMProviderListResponse, error) {
	params := &gen.ListLLMProvidersParams{}

	if limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	if offset > 0 {
		o := int(offset)
		params.Offset = &o
	}

	resp, err := c.apiClient.ListLLMProvidersWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list LLM providers: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from list LLM providers")
	}

	providers := make([]LLMProviderResponse, 0, len(resp.JSON200.List))
	for _, p := range resp.JSON200.List {
		providers = append(providers, *toLLMProviderListItemResponse(&p))
	}

	return &LLMProviderListResponse{
		Providers: providers,
		Total:     int32(resp.JSON200.Count),
		Limit:     int32(resp.JSON200.Pagination.Limit),
		Offset:    int32(resp.JSON200.Pagination.Offset),
		Pagination: PaginationInfo{
			Limit:  int32(resp.JSON200.Pagination.Limit),
			Offset: int32(resp.JSON200.Pagination.Offset),
			Total:  int32(resp.JSON200.Pagination.Total),
		},
	}, nil
}

// UpdateLLMProvider updates an LLM provider in Platform API
func (c *platformAPIClient) UpdateLLMProvider(ctx context.Context, providerID string, req UpdateLLMProviderRequest) (*LLMProviderResponse, error) {
	// Build a JSON merge patch body using only provided fields
	patch := make(map[string]interface{})

	if req.DisplayName != nil {
		// Platform API uses "name" as the human-readable provider name
		patch["name"] = *req.DisplayName
	}
	if req.Description != nil {
		patch["description"] = *req.Description
	}
	if req.Version != nil {
		patch["version"] = *req.Version
	}
	if req.Upstream != nil {
		patch["upstream"] = req.Upstream
	}
	if req.OpenAPI != nil {
		patch["openapi"] = *req.OpenAPI
	}

	if req.AccessControl != nil {
		if ac := buildGenLLMAccessControl(req.AccessControl); ac != nil {
			patch["accessControl"] = ac
		}
	}

	if req.ModelProviders != nil && len(*req.ModelProviders) > 0 {
		modelProviders := make([]gen.LLMModelProvider, 0, len(*req.ModelProviders))
		for _, mp := range *req.ModelProviders {
			genModels := make([]gen.LLMModel, 0, len(mp.Models))
			for _, m := range mp.Models {
				name := m.Name
				model := gen.LLMModel{
					Id: m.ID,
				}
				if name != "" {
					model.Name = &name
				}
				if m.Description != "" {
					desc := m.Description
					model.Description = &desc
				}
				genModels = append(genModels, model)
			}

			provider := gen.LLMModelProvider{
				Id: mp.ID,
			}
			if len(genModels) > 0 {
				provider.Models = &genModels
			}
			if mp.Name != "" {
				name := mp.Name
				provider.Name = &name
			}

			modelProviders = append(modelProviders, provider)
		}
		patch["modelProviders"] = modelProviders
	}

	if req.RateLimiting != nil {
		if rl := buildGenLLMRateLimitingConfig(req.RateLimiting); rl != nil {
			patch["rateLimiting"] = rl
		}
	}

	// If nothing to update, just return the current provider
	if len(patch) == 0 {
		return c.GetLLMProvider(ctx, providerID)
	}

	body, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update LLM provider request: %w", err)
	}

	resp, err := c.apiClient.UpdateLLMProviderWithBodyWithResponse(ctx, providerID, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to update LLM provider: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrLLMProviderNotFound,
		})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from update LLM provider")
	}

	return toLLMProviderResponse(resp.JSON200), nil
}

// DeleteLLMProvider deletes an LLM provider from Platform API
func (c *platformAPIClient) DeleteLLMProvider(ctx context.Context, providerID string) error {
	resp, err := c.apiClient.DeleteLLMProviderWithResponse(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete LLM provider: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrLLMProviderNotFound,
		})
	}

	return nil
}

// GetLLMProviderTemplate retrieves an LLM provider template by ID
func (c *platformAPIClient) GetLLMProviderTemplate(ctx context.Context, templateID string) (*LLMProviderTemplateResponse, error) {
	resp, err := c.apiClient.GetLLMProviderTemplateWithResponse(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider template: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrLLMProviderTemplateNotFound,
		})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from get LLM provider template")
	}

	return toLLMProviderTemplateResponseFromTemplate(resp.JSON200), nil
}

// ListLLMProviderTemplates lists LLM provider templates with pagination
func (c *platformAPIClient) ListLLMProviderTemplates(ctx context.Context, limit, offset int32) (*LLMProviderTemplateListResponse, error) {
	params := &gen.ListLLMProviderTemplatesParams{}

	if limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	if offset > 0 {
		o := int(offset)
		params.Offset = &o
	}

	resp, err := c.apiClient.ListLLMProviderTemplatesWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list LLM provider templates: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from list LLM provider templates")
	}

	templates := make([]LLMProviderTemplateResponse, 0, len(resp.JSON200.List))
	for _, t := range resp.JSON200.List {
		templates = append(templates, *toLLMProviderTemplateResponseFromListItem(&t))
	}

	return &LLMProviderTemplateListResponse{
		Templates: templates,
		Total:     int32(resp.JSON200.Count),
		Limit:     int32(resp.JSON200.Pagination.Limit),
		Offset:    int32(resp.JSON200.Pagination.Offset),
		Pagination: PaginationInfo{
			Limit:  int32(resp.JSON200.Pagination.Limit),
			Offset: int32(resp.JSON200.Pagination.Offset),
			Total:  int32(resp.JSON200.Pagination.Total),
		},
	}, nil
}

// Helper functions to convert between types

func toLLMProviderResponse(p *gen.LLMProvider) *LLMProviderResponse {
	response := &LLMProviderResponse{
		ID:          p.Id,
		Name:        p.Name,
		DisplayName: p.Name,
		Description: utils.StrPointerAsStr(p.Description, ""),
		Version:     p.Version,
		Template:    p.Template,
		Context:     utils.StrPointerAsStr(p.Context, ""),
		VHost:       utils.StrPointerAsStr(p.Vhost, ""),
		OpenAPI:     utils.StrPointerAsStr(p.Openapi, ""),
		CreatedAt:   utils.TimePointerAsTime(p.CreatedAt),
		UpdatedAt:   utils.TimePointerAsTime(p.UpdatedAt),
	}

	// Convert upstream to a generic map
	if upstreamMap, err := genUpstreamToMap(p.Upstream); err == nil && upstreamMap != nil {
		response.Upstream = upstreamMap
	}

	// Convert access control
	response.AccessControl = &LLMAccessControl{
		Mode: string(p.AccessControl.Mode),
	}
	if p.AccessControl.Exceptions != nil {
		response.AccessControl.Exceptions = make([]LLMAccessControlRule, 0, len(*p.AccessControl.Exceptions))
		for _, ex := range *p.AccessControl.Exceptions {
			methods := make([]string, 0, len(ex.Methods))
			for _, m := range ex.Methods {
				methods = append(methods, string(m))
			}
			response.AccessControl.Exceptions = append(response.AccessControl.Exceptions, LLMAccessControlRule{
				Path:    ex.Path,
				Methods: methods,
			})
		}
	}

	// Convert model providers
	if p.ModelProviders != nil {
		response.ModelProviders = make([]LLMModelProvider, 0, len(*p.ModelProviders))
		for _, mp := range *p.ModelProviders {
			models := make([]LLMModel, 0)
			if mp.Models != nil {
				for _, m := range *mp.Models {
					models = append(models, LLMModel{
						ID:          m.Id,
						Name:        utils.StrPointerAsStr(m.Name, ""),
						Description: utils.StrPointerAsStr(m.Description, ""),
					})
				}
			}
			response.ModelProviders = append(response.ModelProviders, LLMModelProvider{
				ID:     mp.Id,
				Name:   utils.StrPointerAsStr(mp.Name, ""),
				Models: models,
			})
		}
	}

	return response
}

func toLLMProviderListItemResponse(p *gen.LLMProviderListItem) *LLMProviderResponse {
	return &LLMProviderResponse{
		ID:          utils.StrPointerAsStr(p.Id, ""),
		Name:        utils.StrPointerAsStr(p.Name, ""),
		DisplayName: utils.StrPointerAsStr(p.Name, ""),
		Description: utils.StrPointerAsStr(p.Description, ""),
		Version:     utils.StrPointerAsStr(p.Version, ""),
		Template:    utils.StrPointerAsStr(p.Template, ""),
		CreatedAt:   utils.TimePointerAsTime(p.CreatedAt),
		UpdatedAt:   utils.TimePointerAsTime(p.UpdatedAt),
	}
}

func toLLMProviderTemplateResponseFromTemplate(t *gen.LLMProviderTemplate) *LLMProviderTemplateResponse {
	response := &LLMProviderTemplateResponse{
		ID:          t.Id,
		Name:        t.Name,
		DisplayName: t.Name,
		Description: utils.StrPointerAsStr(t.Description, ""),
		CreatedAt:   utils.TimePointerAsTime(t.CreatedAt),
		UpdatedAt:   utils.TimePointerAsTime(t.UpdatedAt),
	}

	if t.Metadata != nil && t.Metadata.EndpointUrl != nil {
		response.BaseURL = utils.StrPointerAsStr(t.Metadata.EndpointUrl, "")
	}

	return response
}

func toLLMProviderTemplateResponseFromListItem(t *gen.LLMProviderTemplateListItem) *LLMProviderTemplateResponse {
	return &LLMProviderTemplateResponse{
		ID:          utils.StrPointerAsStr(t.Id, ""),
		Name:        utils.StrPointerAsStr(t.Name, ""),
		DisplayName: utils.StrPointerAsStr(t.Name, ""),
		Description: utils.StrPointerAsStr(t.Description, ""),
		CreatedAt:   utils.TimePointerAsTime(t.CreatedAt),
		UpdatedAt:   utils.TimePointerAsTime(t.UpdatedAt),
	}
}

// mapToGenUpstream converts a generic map representation of upstream into the typed gen.Upstream.
func mapToGenUpstream(data map[string]interface{}) (gen.Upstream, error) {
	var upstream gen.Upstream
	if data == nil {
		return upstream, nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return upstream, err
	}

	if err := json.Unmarshal(raw, &upstream); err != nil {
		return upstream, err
	}

	return upstream, nil
}

// genUpstreamToMap converts a typed gen.Upstream into a generic map representation.
func genUpstreamToMap(upstream gen.Upstream) (map[string]interface{}, error) {
	raw, err := json.Marshal(upstream)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// buildGenLLMAccessControl maps the internal LLMAccessControl representation to the Platform API model.
func buildGenLLMAccessControl(ac *LLMAccessControl) *gen.LLMAccessControl {
	if ac == nil {
		return nil
	}

	result := gen.LLMAccessControl{
		Mode: gen.LLMAccessControlMode(ac.Mode),
	}

	if len(ac.Exceptions) > 0 {
		exceptions := make([]gen.RouteException, 0, len(ac.Exceptions))
		for _, ex := range ac.Exceptions {
			methods := make([]gen.RouteExceptionMethods, 0, len(ex.Methods))
			for _, m := range ex.Methods {
				methods = append(methods, gen.RouteExceptionMethods(m))
			}
			exceptions = append(exceptions, gen.RouteException{
				Path:    ex.Path,
				Methods: methods,
			})
		}
		result.Exceptions = &exceptions
	}

	return &result
}

// buildGenLLMRateLimitingConfig maps the internal LLMRateLimiting representation to the Platform API model.
func buildGenLLMRateLimitingConfig(rl *LLMRateLimiting) *gen.LLMRateLimitingConfig {
	if rl == nil || rl.ProviderLevel == nil || rl.ProviderLevel.Global == nil {
		return nil
	}

	global := rl.ProviderLevel.Global
	config := &gen.LLMRateLimitingConfig{
		ProviderLevel: &gen.RateLimitingScopeConfig{
			Global: &gen.RateLimitingLimitConfig{},
		},
	}

	if global.Request != nil {
		count := int(global.Request.Count)
		enabled := global.Request.Enabled
		requestDim := &gen.RequestRateLimitDimension{
			Count:   &count,
			Enabled: &enabled,
		}
		if global.Request.Reset != nil {
			reset := &gen.RateLimitResetWindow{
				Duration: int(global.Request.Reset.Duration),
				Unit:     gen.RateLimitResetWindowUnit(global.Request.Reset.Unit),
			}
			requestDim.Reset = reset
		}
		config.ProviderLevel.Global.Request = requestDim
	}

	if global.Token != nil {
		count := int(global.Token.Count)
		enabled := global.Token.Enabled
		tokenDim := &gen.TokenRateLimitDimension{
			Count:   &count,
			Enabled: &enabled,
		}
		if global.Token.Reset != nil {
			reset := &gen.RateLimitResetWindow{
				Duration: int(global.Token.Reset.Duration),
				Unit:     gen.RateLimitResetWindowUnit(global.Token.Reset.Unit),
			}
			tokenDim.Reset = reset
		}
		config.ProviderLevel.Global.Token = tokenDim
	}

	return config
}
