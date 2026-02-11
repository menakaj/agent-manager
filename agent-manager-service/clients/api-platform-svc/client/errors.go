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
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorContext provides context-specific error mappings for different operations
type ErrorContext struct {
	NotFoundErr error
	ConflictErr error
}

// PlatformAPIError represents an error response from Platform API
type PlatformAPIError struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// handleErrorResponse processes error responses from Platform API
func handleErrorResponse(statusCode int, body []byte, ctx ErrorContext) error {
	// Try to parse error response
	var apiErr PlatformAPIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
		return fmt.Errorf("Platform API error (HTTP %d): %s - %s", statusCode, apiErr.Message, apiErr.Description)
	}

	// Handle standard HTTP status codes
	switch statusCode {
	case http.StatusNotFound:
		if ctx.NotFoundErr != nil {
			return ctx.NotFoundErr
		}
		return fmt.Errorf("resource not found (HTTP 404)")

	case http.StatusConflict:
		if ctx.ConflictErr != nil {
			return ctx.ConflictErr
		}
		return fmt.Errorf("resource already exists (HTTP 409)")

	case http.StatusBadRequest:
		return fmt.Errorf("invalid request (HTTP 400): %s", string(body))

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized (HTTP 401): invalid or expired token")

	case http.StatusForbidden:
		return fmt.Errorf("forbidden (HTTP 403): insufficient permissions")

	case http.StatusInternalServerError:
		return fmt.Errorf("Platform API internal error (HTTP 500)")

	case http.StatusServiceUnavailable:
		return fmt.Errorf("Platform API unavailable (HTTP 503)")

	default:
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
}
