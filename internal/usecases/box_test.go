package usecases

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestBoxUseCase_BuildBox(t *testing.T) {
	mockTemplate := &mockTemplateAdapter{}
	mockEntry := &mockEntryAdapter{}

	useCase := NewBox(mockTemplate, mockEntry, NewPathUseCase())
	results, err := useCase.BuildBox(context.Background(), "test", "development", "test.json", map[string]string{})

	fmt.Println(results)

	expected := `{"service": "test","ENV_1": "key-test", "ENV_2": "false", "GLOBAL_SERVICE": "xxxxx12345", "domain": "private.io", "version": "1", "missing":""}`

	if err != nil {
		t.Errorf(`Expected %s got: err %s`, expected, err)
	}

	if strings.TrimSpace(results) != strings.TrimSpace(expected) {
		t.Errorf(`Expected %s got: %s`, expected, results)
	}
}
