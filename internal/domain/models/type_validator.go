package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

type TypeValidator struct {
	Name  string `json:"name" example:"email"`
	Regex string `json:"regex" example:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
}

func (tv *TypeValidator) String() string {
	return fmt.Sprintf("Name: %s. Regex: %s", tv.Name, tv.Regex)
}

// Validate validates a value against the type validator's regex
func (tv *TypeValidator) Validate(value string) error {
	matched, err := regexp.MatchString(tv.Regex, value)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	if !matched {
		return fmt.Errorf("value does not match type validator '%s'", tv.Name)
	}
	return nil
}

// Built-in type validators
var (
	TypeValidatorString = TypeValidator{
		Name:  "string",
		Regex: ".*", // Matches any string
	}

	TypeValidatorNumber = TypeValidator{
		Name:  "number",
		Regex: `^-?\d+(\.\d+)?$`, // Matches integers and decimals
	}

	TypeValidatorJSON = TypeValidator{
		Name:  "json",
		Regex: "", // Special case - will use json.Valid
	}

	TypeValidatorURLHTTPS = TypeValidator{
		Name:  "url-https",
		Regex: `^https://[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*(/.*)?$`,
	}

	TypeValidatorURLHTTP = TypeValidator{
		Name:  "url-http",
		Regex: `^http://[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*(/.*)?$`,
	}

	// Map of built-in validators
	BuiltInValidators = map[string]TypeValidator{
		"string":    TypeValidatorString,
		"number":    TypeValidatorNumber,
		"json":      TypeValidatorJSON,
		"url-https": TypeValidatorURLHTTPS,
		"url-http":  TypeValidatorURLHTTP,
	}
)

// ValidateValue validates a value using custom logic for built-in validators
func ValidateValue(validator *TypeValidator, value string) error {
	if validator == nil {
		// Default to string validator
		return nil
	}

	// Special handling for built-in validators
	switch validator.Name {
	case "json":
		return validateJSON(value)
	case "number":
		return validateNumber(value)
	case "url-https":
		return validateURL(value, "https")
	case "url-http":
		return validateURL(value, "http")
	default:
		// Use regex validation
		return validator.Validate(value)
	}
}

func validateJSON(value string) error {
	if !json.Valid([]byte(value)) {
		return fmt.Errorf("value is not valid JSON")
	}
	return nil
}

func validateNumber(value string) error {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("value is not a valid number")
	}
	return nil
}

func validateURL(value string, scheme string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("value is not a valid URL: %w", err)
	}
	if u.Scheme != scheme {
		return fmt.Errorf("URL scheme must be %s", scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}
