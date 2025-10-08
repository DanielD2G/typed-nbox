package models

import (
	"testing"
)

func TestTypeValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator TypeValidator
		value     string
		wantErr   bool
	}{
		{
			name:      "valid string",
			validator: TypeValidatorString,
			value:     "any string value",
			wantErr:   false,
		},
		{
			name:      "valid number - integer",
			validator: TypeValidatorNumber,
			value:     "123",
			wantErr:   false,
		},
		{
			name:      "valid number - decimal",
			validator: TypeValidatorNumber,
			value:     "123.45",
			wantErr:   false,
		},
		{
			name:      "valid number - negative",
			validator: TypeValidatorNumber,
			value:     "-123.45",
			wantErr:   false,
		},
		{
			name:      "invalid number",
			validator: TypeValidatorNumber,
			value:     "not a number",
			wantErr:   true,
		},
		{
			name:      "valid https url",
			validator: TypeValidatorURLHTTPS,
			value:     "https://example.com",
			wantErr:   false,
		},
		{
			name:      "valid https url with path",
			validator: TypeValidatorURLHTTPS,
			value:     "https://example.com/path/to/resource",
			wantErr:   false,
		},
		{
			name:      "invalid https url - http scheme",
			validator: TypeValidatorURLHTTPS,
			value:     "http://example.com",
			wantErr:   true,
		},
		{
			name:      "invalid https url - no scheme",
			validator: TypeValidatorURLHTTPS,
			value:     "example.com",
			wantErr:   true,
		},
		{
			name:      "valid http url",
			validator: TypeValidatorURLHTTP,
			value:     "http://example.com",
			wantErr:   false,
		},
		{
			name:      "invalid http url - https scheme",
			validator: TypeValidatorURLHTTP,
			value:     "https://example.com",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypeValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		name      string
		validator *TypeValidator
		value     string
		wantErr   bool
	}{
		{
			name:      "nil validator - should pass",
			validator: nil,
			value:     "any value",
			wantErr:   false,
		},
		{
			name:      "valid json object",
			validator: &TypeValidatorJSON,
			value:     `{"key": "value"}`,
			wantErr:   false,
		},
		{
			name:      "valid json array",
			validator: &TypeValidatorJSON,
			value:     `["item1", "item2"]`,
			wantErr:   false,
		},
		{
			name:      "invalid json",
			validator: &TypeValidatorJSON,
			value:     `{invalid json}`,
			wantErr:   true,
		},
		{
			name:      "valid number via ValidateValue",
			validator: &TypeValidatorNumber,
			value:     "42",
			wantErr:   false,
		},
		{
			name:      "invalid number via ValidateValue",
			validator: &TypeValidatorNumber,
			value:     "abc",
			wantErr:   true,
		},
		{
			name:      "valid url-https via ValidateValue",
			validator: &TypeValidatorURLHTTPS,
			value:     "https://google.com",
			wantErr:   false,
		},
		{
			name:      "invalid url-https - missing host",
			validator: &TypeValidatorURLHTTPS,
			value:     "https://",
			wantErr:   true,
		},
		{
			name:      "valid url-http via ValidateValue",
			validator: &TypeValidatorURLHTTP,
			value:     "http://localhost:8080",
			wantErr:   false,
		},
		{
			name:      "custom regex validator - email pattern",
			validator: &TypeValidator{Name: "email", Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
			value:     "test@example.com",
			wantErr:   false,
		},
		{
			name:      "custom regex validator - invalid email",
			validator: &TypeValidator{Name: "email", Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
			value:     "invalid-email",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValue(tt.validator, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid empty object",
			value:   `{}`,
			wantErr: false,
		},
		{
			name:    "valid nested object",
			value:   `{"user": {"name": "John", "age": 30}}`,
			wantErr: false,
		},
		{
			name:    "invalid json - missing quotes",
			value:   `{key: value}`,
			wantErr: true,
		},
		{
			name:    "invalid json - trailing comma",
			value:   `{"key": "value",}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJSON(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNumber(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid integer",
			value:   "42",
			wantErr: false,
		},
		{
			name:    "valid negative integer",
			value:   "-42",
			wantErr: false,
		},
		{
			name:    "valid float",
			value:   "3.14159",
			wantErr: false,
		},
		{
			name:    "valid scientific notation",
			value:   "1.23e10",
			wantErr: false,
		},
		{
			name:    "invalid - text",
			value:   "not a number",
			wantErr: true,
		},
		{
			name:    "invalid - empty string",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNumber(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		scheme  string
		wantErr bool
	}{
		{
			name:    "valid https url",
			value:   "https://example.com",
			scheme:  "https",
			wantErr: false,
		},
		{
			name:    "valid http url with port",
			value:   "http://localhost:8080",
			scheme:  "http",
			wantErr: false,
		},
		{
			name:    "invalid - wrong scheme",
			value:   "https://example.com",
			scheme:  "http",
			wantErr: true,
		},
		{
			name:    "invalid - no host",
			value:   "https://",
			scheme:  "https",
			wantErr: true,
		},
		{
			name:    "invalid - malformed url",
			value:   "not a url",
			scheme:  "https",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.value, tt.scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuiltInValidators(t *testing.T) {
	expectedValidators := []string{"string", "number", "json", "url-https", "url-http"}

	for _, name := range expectedValidators {
		t.Run("check_builtin_"+name, func(t *testing.T) {
			_, exists := BuiltInValidators[name]
			if !exists {
				t.Errorf("Built-in validator '%s' not found in BuiltInValidators map", name)
			}
		})
	}

	if len(BuiltInValidators) != len(expectedValidators) {
		t.Errorf("Expected %d built-in validators, got %d", len(expectedValidators), len(BuiltInValidators))
	}
}

func TestTypeValidator_String(t *testing.T) {
	validator := TypeValidator{
		Name:  "test-validator",
		Regex: "^test.*$",
	}

	result := validator.String()
	expected := "Name: test-validator. Regex: ^test.*$"

	if result != expected {
		t.Errorf("TypeValidator.String() = %v, want %v", result, expected)
	}
}
