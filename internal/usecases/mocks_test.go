package usecases

import (
	"context"
	"encoding/json"
	"nbox/internal/domain/models"
	"nbox/internal/domain/models/operations"
)

type mockTemplateAdapter struct {
}

type mockEntryAdapter struct {
}

func (m *mockEntryAdapter) Upsert(_ context.Context, _ []models.Entry) operations.Results {
	return nil
}

func (m *mockEntryAdapter) Retrieve(_ context.Context, _ string) (*models.Entry, error) {
	return nil, nil
}

func (m *mockEntryAdapter) List(_ context.Context, _ string) ([]models.Entry, error) {
	text := `[
		{ "path": "widget-x/development", "key": "key", "value": "key-test", "secure": false },
		{ "path": "widget-x/development", "key": "debug", "value": "false", "secure": false },
		{ "path": "widget-x", "key": "sentry", "value": "xxxxx12345", "secure": false },
		{ "path": " ", "key": "private-domain", "value": "private.io", "secure": false }
	]`
	var entries []models.Entry
	_ = json.Unmarshal([]byte(text), &entries)
	return entries, nil
}

func (m *mockEntryAdapter) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockEntryAdapter) Tracking(_ context.Context, _ string) ([]models.Tracking, error) {
	return nil, nil
}

func (m *mockTemplateAdapter) UpsertBox(ctx context.Context, box *models.Box) []string {
	return nil
}

func (m *mockTemplateAdapter) BoxExists(ctx context.Context, service string, stage string, template string) (bool, error) {
	return false, nil
}

func (m *mockTemplateAdapter) RetrieveBox(ctx context.Context, service string, stage string, template string) ([]byte, error) {
	text := `{"service": ":service","ENV_1": "{{ widget-x/:stage/key }}", "ENV_2": "{{widget-x/development/debug}}", "GLOBAL_SERVICE": "{{widget-x/sentry}}", "domain": "{{private-domain}}", "version": "1", "missing":"{{missing}}"}`
	return []byte(text), nil
}

func (m *mockTemplateAdapter) List(ctx context.Context) ([]models.Box, error) {
	return nil, nil
}
