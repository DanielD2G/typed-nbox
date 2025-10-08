package usecases

import (
	"context"
	"errors"
	"nbox/internal/application"
	"nbox/internal/domain/models"
	"nbox/internal/domain/models/operations"
	"testing"
)

// Mock adapters for testing with custom upsert function
type mockEntryAdapterWithUpsert struct {
	mockEntryAdapter
	upsertFunc func(ctx context.Context, entries []models.Entry) operations.Results
}

func (m *mockEntryAdapterWithUpsert) Upsert(ctx context.Context, entries []models.Entry) operations.Results {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, entries)
	}
	results := make(operations.Results)
	for _, entry := range entries {
		results[entry.Key] = operations.Result{
			Key:   entry.Key,
			Type:  operations.Updated,
			Error: nil,
		}
	}
	return results
}

type mockSecretAdapter struct {
	upsertFunc func(ctx context.Context, entries []models.Entry) operations.Results
}

func (m *mockSecretAdapter) Upsert(ctx context.Context, entries []models.Entry) operations.Results {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, entries)
	}
	results := make(operations.Results)
	for _, entry := range entries {
		results[entry.Key] = operations.Result{
			Key:   entry.Key,
			Type:  operations.Updated,
			Error: nil,
		}
	}
	return results
}

func (m *mockSecretAdapter) RetrieveSecretValue(ctx context.Context, key string) (*models.Entry, error) {
	return nil, nil
}

type mockTypeValidatorAdapter struct {
	retrieveFunc func(ctx context.Context, name string) (*models.TypeValidator, error)
	upsertFunc   func(ctx context.Context, validator models.TypeValidator) error
	listFunc     func(ctx context.Context) ([]models.TypeValidator, error)
	deleteFunc   func(ctx context.Context, name string) error
}

func (m *mockTypeValidatorAdapter) Retrieve(ctx context.Context, name string) (*models.TypeValidator, error) {
	if m.retrieveFunc != nil {
		return m.retrieveFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockTypeValidatorAdapter) Upsert(ctx context.Context, validator models.TypeValidator) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, validator)
	}
	return nil
}

func (m *mockTypeValidatorAdapter) List(ctx context.Context) ([]models.TypeValidator, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockTypeValidatorAdapter) Delete(ctx context.Context, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, name)
	}
	return nil
}

func TestEntryUseCase_Upsert_WithValidation(t *testing.T) {
	tests := []struct {
		name                 string
		entries              []models.Entry
		typeValidatorAdapter *mockTypeValidatorAdapter
		wantErrorForKey      map[string]bool
		wantSuccessCount     int
	}{
		{
			name: "valid entry with built-in number validator",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             "123",
					TypeValidatorName: "number",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{},
			wantSuccessCount:     1,
		},
		{
			name: "invalid entry with built-in number validator",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             "not-a-number",
					TypeValidatorName: "number",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{"test/key": true},
			wantSuccessCount:     0,
		},
		{
			name: "valid entry with built-in json validator",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             `{"valid": "json"}`,
					TypeValidatorName: "json",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{},
			wantSuccessCount:     1,
		},
		{
			name: "invalid entry with built-in json validator",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             `{invalid json}`,
					TypeValidatorName: "json",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{"test/key": true},
			wantSuccessCount:     0,
		},
		{
			name: "valid entry with built-in url-https validator",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             "https://example.com",
					TypeValidatorName: "url-https",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{},
			wantSuccessCount:     1,
		},
		{
			name: "invalid entry with built-in url-https validator",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             "http://example.com",
					TypeValidatorName: "url-https",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{"test/key": true},
			wantSuccessCount:     0,
		},
		{
			name: "entry without validator - should pass",
			entries: []models.Entry{
				{
					Key:   "test/key",
					Value: "any value",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{},
			wantSuccessCount:     1,
		},
		{
			name: "custom validator from adapter",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             "test@example.com",
					TypeValidatorName: "email",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{
				retrieveFunc: func(ctx context.Context, name string) (*models.TypeValidator, error) {
					if name == "email" {
						return &models.TypeValidator{
							Name:  "email",
							Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
						}, nil
					}
					return nil, nil
				},
			},
			wantErrorForKey:  map[string]bool{},
			wantSuccessCount: 1,
		},
		{
			name: "custom validator not found",
			entries: []models.Entry{
				{
					Key:               "test/key",
					Value:             "some value",
					TypeValidatorName: "nonexistent",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{
				retrieveFunc: func(ctx context.Context, name string) (*models.TypeValidator, error) {
					return nil, errors.New("validator not found")
				},
			},
			wantErrorForKey:  map[string]bool{"test/key": true},
			wantSuccessCount: 0,
		},
		{
			name: "multiple entries with mixed validation results",
			entries: []models.Entry{
				{
					Key:               "test/key1",
					Value:             "123",
					TypeValidatorName: "number",
				},
				{
					Key:               "test/key2",
					Value:             "not-a-number",
					TypeValidatorName: "number",
				},
				{
					Key:   "test/key3",
					Value: "no validator",
				},
			},
			typeValidatorAdapter: &mockTypeValidatorAdapter{},
			wantErrorForKey:      map[string]bool{"test/key2": true},
			wantSuccessCount:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entryAdapter := &mockEntryAdapterWithUpsert{}
			secretAdapter := &mockSecretAdapter{}
			config := &application.Config{
				ParameterShortArn: true,
			}

			useCase := NewEntryUseCase(
				entryAdapter,
				secretAdapter,
				tt.typeValidatorAdapter,
				config,
			)

			results := useCase.Upsert(context.Background(), tt.entries)

			// Count successful operations
			successCount := 0
			for _, result := range results {
				if result.Error == nil {
					successCount++
				}
			}

			if successCount != tt.wantSuccessCount {
				t.Errorf("Expected %d successful operations, got %d", tt.wantSuccessCount, successCount)
			}

			// Check specific error expectations
			for key, shouldHaveError := range tt.wantErrorForKey {
				var found bool
				var result operations.Result
				for _, r := range results {
					if r.Key == key {
						result = r
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected result for key %s, but not found", key)
					continue
				}

				hasError := result.Error != nil
				if hasError != shouldHaveError {
					t.Errorf("Key %s: expected error=%v, got error=%v (actual error: %v)",
						key, shouldHaveError, hasError, result.Error)
				}
			}
		})
	}
}

func TestEntryUseCase_Upsert_SecureEntries(t *testing.T) {
	entryAdapter := &mockEntryAdapterWithUpsert{}
	secretAdapter := &mockSecretAdapter{
		upsertFunc: func(ctx context.Context, entries []models.Entry) operations.Results {
			results := make(operations.Results)
			for _, entry := range entries {
				results[entry.Key] = operations.Result{
					Key:   entry.Key,
					Type:  operations.Updated,
					Error: nil,
				}
			}
			return results
		},
	}
	typeValidatorAdapter := &mockTypeValidatorAdapter{}
	config := &application.Config{
		ParameterShortArn: true,
		RegionName:        "us-east-1",
		AccountId:         "123456789",
	}

	useCase := NewEntryUseCase(
		entryAdapter,
		secretAdapter,
		typeValidatorAdapter,
		config,
	)

	entries := []models.Entry{
		{
			Key:    "test/secure-key",
			Value:  "secret-value",
			Secure: true,
		},
	}

	results := useCase.Upsert(context.Background(), entries)

	if len(results) == 0 {
		t.Error("Expected results, got none")
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Expected no error for secure entry, got: %v", result.Error)
		}
	}
}

func TestEntryUseCase_Upsert_ValidationBeforeSecure(t *testing.T) {
	entryAdapter := &mockEntryAdapterWithUpsert{}
	secretAdapterCalled := false
	secretAdapter := &mockSecretAdapter{
		upsertFunc: func(ctx context.Context, entries []models.Entry) operations.Results {
			secretAdapterCalled = true
			return make(operations.Results)
		},
	}
	typeValidatorAdapter := &mockTypeValidatorAdapter{}
	config := &application.Config{
		ParameterShortArn: true,
	}

	useCase := NewEntryUseCase(
		entryAdapter,
		secretAdapter,
		typeValidatorAdapter,
		config,
	)

	entries := []models.Entry{
		{
			Key:               "test/secure-key",
			Value:             "not-a-number",
			Secure:            true,
			TypeValidatorName: "number",
		},
	}

	results := useCase.Upsert(context.Background(), entries)

	// Should have validation error
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	for _, result := range results {
		if result.Error == nil {
			t.Error("Expected validation error, got nil")
		}
	}

	// Secret adapter should not be called because validation failed
	if secretAdapterCalled {
		t.Error("Secret adapter should not be called when validation fails")
	}
}
