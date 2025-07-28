package usecases

import (
	"context"
	"encoding/json"
	"github.com/norlis/httpgate/pkg/adapter/apidriven/middleware"
	"nbox/internal/application"
	"nbox/internal/domain"
	"nbox/internal/domain/models"
	"nbox/internal/domain/models/operations"
	"time"
)

type entryUseCaseWithEvents struct {
	wrappedUseCase domain.EntryUseCase
	notifier       domain.EventNotifier
}

func NewEntryUseCaseWithEvents(uc domain.EntryUseCase, notifier domain.EventNotifier) domain.EntryUseCase {
	return &entryUseCaseWithEvents{
		wrappedUseCase: uc,
		notifier:       notifier,
	}
}

func (d *entryUseCaseWithEvents) Upsert(ctx context.Context, entries []models.Entry) []operations.Result {
	results := d.wrappedUseCase.Upsert(ctx, entries)

	id := middleware.TraceIdFromContext(ctx)
	user, ok := application.UserFromContext(ctx)

	payload, _ := json.Marshal(results)
	updatedBy := "ghost"

	if ok {
		updatedBy = user.Name
	}
	event := domain.Event[json.RawMessage]{
		Username:      updatedBy,
		TransactionId: id,
		Type:          domain.EventEntryActions,
		Timestamp:     time.Now().UTC(),
		Payload:       payload,
	}
	d.notifier.Dispatch(ctx, event)
	return results
}
