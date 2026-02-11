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
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/wso2/ai-agent-management-platform/agent-manager-service/clients/api-platform-svc/gen"
	"github.com/wso2/ai-agent-management-platform/agent-manager-service/utils"
)

// CreateGateway creates a new gateway in Platform API
func (c *platformAPIClient) CreateGateway(ctx context.Context, req CreateGatewayRequest) (*GatewayResponse, error) {
	// Map functionality type to enum
	var funcType gen.CreateGatewayRequestFunctionalityType
	switch req.FunctionalityType {
	case "ai":
		funcType = gen.CreateGatewayRequestFunctionalityTypeAi
	case "regular":
		funcType = gen.CreateGatewayRequestFunctionalityTypeRegular
	case "event":
		funcType = gen.CreateGatewayRequestFunctionalityTypeEvent
	default:
		funcType = gen.CreateGatewayRequestFunctionalityTypeAi // Default to AI
	}

	apiReq := gen.CreateGatewayJSONRequestBody{
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Vhost:             req.VHost,
		FunctionalityType: funcType,
	}

	// Add optional fields
	if req.Description != "" {
		apiReq.Description = &req.Description
	}
	if req.IsCritical {
		apiReq.IsCritical = &req.IsCritical
	}
	if len(req.Properties) > 0 {
		apiReq.Properties = &req.Properties
	}

	resp, err := c.apiClient.CreateGatewayWithResponse(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			ConflictErr: utils.ErrGatewayAlreadyExists,
		})
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("empty response from create gateway")
	}

	return toGatewayResponse(resp.JSON201), nil
}

// GetGateway retrieves a gateway by ID from Platform API
func (c *platformAPIClient) GetGateway(ctx context.Context, gatewayID string) (*GatewayResponse, error) {
	// Parse gateway ID as UUID
	gwUUID, err := uuid.Parse(gatewayID)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway ID: %w", err)
	}

	resp, err := c.apiClient.GetGatewayWithResponse(ctx, gwUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrGatewayNotFound,
		})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from get gateway")
	}

	return toGatewayResponse(resp.JSON200), nil
}

// ListGateways lists gateways with optional filters
func (c *platformAPIClient) ListGateways(ctx context.Context, filters GatewayFilters) (*GatewayListResponse, error) {
	resp, err := c.apiClient.ListGatewaysWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list gateways: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from list gateways")
	}

	gateways := make([]GatewayResponse, 0, len(resp.JSON200.List))
	for _, gw := range resp.JSON200.List {
		gateways = append(gateways, *toGatewayResponse(&gw))
	}

	return &GatewayListResponse{
		Gateways: gateways,
		Total:    int32(resp.JSON200.Count),
		Limit:    int32(resp.JSON200.Pagination.Limit),
		Offset:   int32(resp.JSON200.Pagination.Offset),
		Pagination: PaginationInfo{
			Limit:  int32(resp.JSON200.Pagination.Limit),
			Offset: int32(resp.JSON200.Pagination.Offset),
			Total:  int32(resp.JSON200.Pagination.Total),
		},
	}, nil
}

// UpdateGateway updates a gateway in Platform API
func (c *platformAPIClient) UpdateGateway(ctx context.Context, gatewayID string, req UpdateGatewayRequest) (*GatewayResponse, error) {
	// Parse gateway ID as UUID
	gwUUID, err := uuid.Parse(gatewayID)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway ID: %w", err)
	}

	apiReq := gen.UpdateGatewayJSONRequestBody{}

	// Add optional fields
	if req.DisplayName != "" {
		apiReq.DisplayName = &req.DisplayName
	}
	if req.Description != "" {
		apiReq.Description = &req.Description
	}
	if req.IsCritical != nil {
		apiReq.IsCritical = req.IsCritical
	}
	if len(req.Properties) > 0 {
		apiReq.Properties = &req.Properties
	}

	resp, err := c.apiClient.UpdateGatewayWithResponse(ctx, gwUUID, apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update gateway: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrGatewayNotFound,
		})
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response from update gateway")
	}

	return toGatewayResponse(resp.JSON200), nil
}

// DeleteGateway deletes a gateway from Platform API
func (c *platformAPIClient) DeleteGateway(ctx context.Context, gatewayID string) error {
	// Parse gateway ID as UUID
	gwUUID, err := uuid.Parse(gatewayID)
	if err != nil {
		return fmt.Errorf("invalid gateway ID: %w", err)
	}

	resp, err := c.apiClient.DeleteGatewayWithResponse(ctx, gwUUID)
	if err != nil {
		return fmt.Errorf("failed to delete gateway: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrGatewayNotFound,
		})
	}

	return nil
}

// RotateGatewayToken rotates the authentication token for a gateway
func (c *platformAPIClient) RotateGatewayToken(ctx context.Context, gatewayID string) (*TokenRotationResponse, error) {
	// Parse gateway ID as UUID
	gwUUID, err := uuid.Parse(gatewayID)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway ID: %w", err)
	}

	resp, err := c.apiClient.RotateGatewayTokenWithResponse(ctx, gwUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate gateway token: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return nil, handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrGatewayNotFound,
		})
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("empty response from rotate gateway token")
	}

	token := resp.JSON201
	result := &TokenRotationResponse{
		TokenID:   utils.UUIDPointerAsStr(token.Id),
		Token:     utils.StrPointerAsStr(token.Token, ""),
		CreatedAt: utils.TimePointerAsTime(token.CreatedAt),
	}

	return result, nil
}

// RevokeGatewayToken revokes a specific gateway token
func (c *platformAPIClient) RevokeGatewayToken(ctx context.Context, gatewayID, tokenID string) error {
	// Parse UUIDs
	gwUUID, err := uuid.Parse(gatewayID)
	if err != nil {
		return fmt.Errorf("invalid gateway ID: %w", err)
	}

	tokenUUID, err := uuid.Parse(tokenID)
	if err != nil {
		return fmt.Errorf("invalid token ID: %w", err)
	}

	resp, err := c.apiClient.RevokeGatewayTokenWithResponse(ctx, gwUUID, tokenUUID)
	if err != nil {
		return fmt.Errorf("failed to revoke gateway token: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return handleErrorResponse(resp.StatusCode(), resp.Body, ErrorContext{
			NotFoundErr: utils.ErrGatewayNotFound,
		})
	}

	return nil
}

// toGatewayResponse converts Platform API gateway response to internal type
func toGatewayResponse(gw *gen.GatewayResponse) *GatewayResponse {
	response := &GatewayResponse{
		ID:             utils.UUIDPointerAsStr(gw.Id),
		OrganizationID: utils.UUIDPointerAsStr(gw.OrganizationId),
		Name:           utils.StrPointerAsStr(gw.Name, ""),
		DisplayName:    utils.StrPointerAsStr(gw.DisplayName, ""),
		Description:    utils.StrPointerAsStr(gw.Description, ""),
		VHost:          utils.StrPointerAsStr(gw.Vhost, ""),
		IsCritical:     utils.BoolPointerAsBool(gw.IsCritical, false),
		IsActive:       utils.BoolPointerAsBool(gw.IsActive, false),
		CreatedAt:      utils.TimePointerAsTime(gw.CreatedAt),
		UpdatedAt:      utils.TimePointerAsTime(gw.UpdatedAt),
	}

	// Map functionality type
	if gw.FunctionalityType != nil {
		response.FunctionalityType = string(*gw.FunctionalityType)
	}

	// Copy properties
	if gw.Properties != nil {
		response.Properties = *gw.Properties
	}

	return response
}
