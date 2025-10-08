package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"nbox/internal/domain/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/norlis/httpgate/pkg/adapter/apidriven/presenters"
	"go.uber.org/zap"
)

// Mock TypeValidatorAdapter for testing
type mockTypeValidatorAdapter struct {
	upsertFunc   func(ctx context.Context, validator models.TypeValidator) error
	retrieveFunc func(ctx context.Context, name string) (*models.TypeValidator, error)
	listFunc     func(ctx context.Context) ([]models.TypeValidator, error)
	deleteFunc   func(ctx context.Context, name string) error
}

func (m *mockTypeValidatorAdapter) Upsert(ctx context.Context, validator models.TypeValidator) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, validator)
	}
	return nil
}

func (m *mockTypeValidatorAdapter) Retrieve(ctx context.Context, name string) (*models.TypeValidator, error) {
	if m.retrieveFunc != nil {
		return m.retrieveFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockTypeValidatorAdapter) List(ctx context.Context) ([]models.TypeValidator, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []models.TypeValidator{}, nil
}

func (m *mockTypeValidatorAdapter) Delete(ctx context.Context, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, name)
	}
	return nil
}

func TestTypeValidatorHandler_Upsert(t *testing.T) {
	tests := []struct {
		name           string
		validator      models.TypeValidator
		adapterFunc    func(ctx context.Context, validator models.TypeValidator) error
		expectedStatus int
	}{
		{
			name: "successful upsert",
			validator: models.TypeValidator{
				Name:  "email",
				Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			adapterFunc: func(ctx context.Context, validator models.TypeValidator) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "adapter error",
			validator: models.TypeValidator{
				Name:  "test",
				Regex: ".*",
			},
			adapterFunc: func(ctx context.Context, validator models.TypeValidator) error {
				return errors.New("adapter error")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &mockTypeValidatorAdapter{
				upsertFunc: tt.adapterFunc,
			}
			render := presenters.NewPresenters(zap.NewNop())
			handler := NewTypeValidatorHandler(adapter, render)

			body, _ := json.Marshal(tt.validator)
			req := httptest.NewRequest(http.MethodPost, "/api/type-validator", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			handler.Upsert(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestTypeValidatorHandler_Upsert_InvalidJSON(t *testing.T) {
	adapter := &mockTypeValidatorAdapter{}
	render := presenters.NewPresenters(zap.NewNop())
	handler := NewTypeValidatorHandler(adapter, render)

	req := httptest.NewRequest(http.MethodPost, "/api/type-validator", bytes.NewBufferString("{invalid json}"))
	rec := httptest.NewRecorder()

	handler.Upsert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestTypeValidatorHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		adapterFunc    func(ctx context.Context) ([]models.TypeValidator, error)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful list",
			adapterFunc: func(ctx context.Context) ([]models.TypeValidator, error) {
				return []models.TypeValidator{
					{Name: "email", Regex: ".*"},
					{Name: "phone", Regex: ".*"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "empty list",
			adapterFunc: func(ctx context.Context) ([]models.TypeValidator, error) {
				return []models.TypeValidator{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "adapter error",
			adapterFunc: func(ctx context.Context) ([]models.TypeValidator, error) {
				return nil, errors.New("adapter error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &mockTypeValidatorAdapter{
				listFunc: tt.adapterFunc,
			}
			render := presenters.NewPresenters(zap.NewNop())
			handler := NewTypeValidatorHandler(adapter, render)

			req := httptest.NewRequest(http.MethodGet, "/api/type-validator", nil)
			rec := httptest.NewRecorder()

			handler.List(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK && tt.expectedCount > 0 {
				var result []models.TypeValidator
				if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				} else if len(result) != tt.expectedCount {
					t.Errorf("Expected %d validators, got %d", tt.expectedCount, len(result))
				}
			}
		})
	}
}

func TestTypeValidatorHandler_GetByName(t *testing.T) {
	tests := []struct {
		name           string
		queryName      string
		adapterFunc    func(ctx context.Context, name string) (*models.TypeValidator, error)
		expectedStatus int
	}{
		{
			name:      "successful retrieve",
			queryName: "email",
			adapterFunc: func(ctx context.Context, name string) (*models.TypeValidator, error) {
				return &models.TypeValidator{
					Name:  "email",
					Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "validator not found",
			queryName: "nonexistent",
			adapterFunc: func(ctx context.Context, name string) (*models.TypeValidator, error) {
				return nil, nil
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing name parameter",
			queryName:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "adapter error",
			queryName: "test",
			adapterFunc: func(ctx context.Context, name string) (*models.TypeValidator, error) {
				return nil, errors.New("adapter error")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &mockTypeValidatorAdapter{
				retrieveFunc: tt.adapterFunc,
			}
			render := presenters.NewPresenters(zap.NewNop())
			handler := NewTypeValidatorHandler(adapter, render)

			url := "/api/type-validator/name"
			if tt.queryName != "" {
				url += "?name=" + tt.queryName
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			handler.GetByName(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestTypeValidatorHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		queryName      string
		adapterFunc    func(ctx context.Context, name string) error
		expectedStatus int
	}{
		{
			name:      "successful delete",
			queryName: "email",
			adapterFunc: func(ctx context.Context, name string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing name parameter",
			queryName:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "adapter error - built-in validator",
			queryName: "string",
			adapterFunc: func(ctx context.Context, name string) error {
				return errors.New("cannot delete built-in type validator")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "adapter error - general",
			queryName: "test",
			adapterFunc: func(ctx context.Context, name string) error {
				return errors.New("adapter error")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &mockTypeValidatorAdapter{
				deleteFunc: tt.adapterFunc,
			}
			render := presenters.NewPresenters(zap.NewNop())
			handler := NewTypeValidatorHandler(adapter, render)

			url := "/api/type-validator/name"
			if tt.queryName != "" {
				url += "?name=" + tt.queryName
			}

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			rec := httptest.NewRecorder()

			handler.Delete(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var result map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				} else if result["message"] != "ok" {
					t.Errorf("Expected message 'ok', got '%s'", result["message"])
				}
			}
		})
	}
}