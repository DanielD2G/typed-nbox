package usecases

import (
	"context"
	"fmt"
	"nbox/internal/application"
	"nbox/internal/domain"
	"nbox/internal/domain/models"
	"nbox/internal/domain/models/operations"
	"strings"
)

type EntryUseCase struct {
	entryAdapter         domain.EntryAdapter
	secretAdapter        domain.SecretAdapter
	typeValidatorAdapter domain.TypeValidatorAdapter
	config               *application.Config
}

func NewEntryUseCase(
	entryAdapter domain.EntryAdapter,
	secretAdapter domain.SecretAdapter,
	typeValidatorAdapter domain.TypeValidatorAdapter,
	config *application.Config,
) domain.EntryUseCase {
	return &EntryUseCase{
		entryAdapter:         entryAdapter,
		secretAdapter:        secretAdapter,
		typeValidatorAdapter: typeValidatorAdapter,
		config:               config,
	}
}

// Upsert
// ARN arn:aws:ssm:<REGION_NAME>:<ACCOUNT_ID>:parameter/<parameter-name>
func (e *EntryUseCase) Upsert(ctx context.Context, entries []models.Entry) []operations.Result {
	var results []operations.Result

	// Validate entries with type validators
	validatedEntries := make([]models.Entry, 0)
	for _, entry := range entries {
		// Check if entry already exists to prevent type validator changes
		existingEntry, err := e.entryAdapter.Retrieve(ctx, entry.Key)
		if err == nil && existingEntry != nil {
			// Entry exists, check if type validator is being changed
			if existingEntry.TypeValidatorName != entry.TypeValidatorName {
				results = append(results, operations.Result{
					Key:   entry.Key,
					Type:  operations.Error,
					Error: fmt.Errorf("cannot change type validator for existing key '%s' from '%s' to '%s'. Delete and recreate the entry to change type",
						entry.Key, existingEntry.TypeValidatorName, entry.TypeValidatorName),
				})
				continue
			}
		}

		// Validate type if type_validator_name is provided
		if entry.TypeValidatorName != "" {
			// Check if it's a built-in validator
			validator, exists := models.BuiltInValidators[entry.TypeValidatorName]
			if !exists {
				// Try to retrieve custom validator
				customValidator, err := e.typeValidatorAdapter.Retrieve(ctx, entry.TypeValidatorName)
				if err != nil || customValidator == nil {
					results = append(results, operations.Result{
						Key:   entry.Key,
						Type:  operations.Error,
						Error: fmt.Errorf("type validator '%s' not found", entry.TypeValidatorName),
					})
					continue
				}
				validator = *customValidator
			}

			// Validate the value
			if err := models.ValidateValue(&validator, entry.Value); err != nil {
				results = append(results, operations.Result{
					Key:   entry.Key,
					Type:  operations.Error,
					Error: fmt.Errorf("validation failed for key '%s': %w", entry.Key, err),
				})
				continue
			}
		}

		validatedEntries = append(validatedEntries, entry)
	}

	secrets := make([]models.Entry, 0)
	for _, entry := range validatedEntries {
		if entry.Secure {
			secrets = append(secrets, entry)
		}
	}

	secureResults := e.secretAdapter.Upsert(ctx, secrets)

	for i, entry := range validatedEntries {
		if entry.Secure {
			err := secureResults[entry.Key].Error
			validatedEntries[i].Value = ""

			if err != nil {
				continue
			}

			key := cleanedKey(entry.Key)
			validatedEntries[i].Value = e.GetParameterArn(key)
		}
	}

	updated := e.entryAdapter.Upsert(ctx, validatedEntries)

	for k, v := range secureResults {
		updated[k] = v
	}

	for _, v := range updated {
		results = append(results, v)
	}

	return results
}

func (e *EntryUseCase) GetParameterArn(key string) string {
	if e.config.ParameterShortArn && !strings.HasPrefix(key, "/") {
		return "/" + key
	}

	if e.config.ParameterShortArn {
		return key
	}

	return fmt.Sprintf(
		"arn:aws:ssm:%s:%s:parameter/%s", e.config.RegionName, e.config.AccountId, cleanedKey(key),
	)
}

func cleanedKey(key string) string {
	return strings.TrimPrefix(key, "/")
}
