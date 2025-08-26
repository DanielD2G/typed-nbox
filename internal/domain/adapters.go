package domain

import (
	"context"
	"encoding/json"
	"nbox/internal/domain/models"
	"nbox/internal/domain/models/operations"
)

//var (
//	ErrUpsertSecret = errors.New("error al guardar el secreto en Parameter Store")
//	ErrAddTags      = errors.New("error al añadir etiquetas al recurso de Parameter Store")
//)

type EntryUseCase interface {
	Upsert(ctx context.Context, entries []models.Entry) []operations.Result
}

// TemplateAdapter store templates
type TemplateAdapter interface {
	UpsertBox(ctx context.Context, box *models.Box) []string
	BoxExists(ctx context.Context, service string, stage string, template string) (bool, error)
	RetrieveBox(ctx context.Context, service string, stage string, template string) ([]byte, error)
	List(ctx context.Context) ([]models.Box, error)
}

// EntryAdapter vars backend operations
type EntryAdapter interface {
	Upsert(ctx context.Context, entries []models.Entry) operations.Results
	Retrieve(ctx context.Context, key string) (*models.Entry, error)
	List(ctx context.Context, prefix string) ([]models.Entry, error)
	Delete(ctx context.Context, key string) error
	Tracking(ctx context.Context, key string) ([]models.Tracking, error)
}

// SecretAdapter vars encrypt
type SecretAdapter interface {
	Upsert(ctx context.Context, entries []models.Entry) operations.Results
	RetrieveSecretValue(ctx context.Context, key string) (*models.Entry, error)
}

type EventNotifier interface {
	Dispatch(ctx context.Context, event Event[json.RawMessage])
}

type WebhookRepository interface {
	FindByEventType(ctx context.Context, eventType EventType) ([]Webhook, error)
	// TODO
	// Create(ctx context.Context, webhook Webhook) error
	// Delete(ctx context.Context, webhookID string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event[json.RawMessage]) error
}
