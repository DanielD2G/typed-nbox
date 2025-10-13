package exporter

import (
	"fmt"
	"nbox/internal/domain"
	"nbox/internal/domain/models"
	"strings"
)

type DotEnvExporter struct{}

func NewDotEnvExporter() *DotEnvExporter {
	return &DotEnvExporter{}
}

// Export exporta entries a formato .env (key=value simple)
func (e *DotEnvExporter) Export(entries []models.Entry) ([]byte, error) {
	var builder strings.Builder

	for _, entry := range entries {
		key := domain.ConvertToEnvVarName(entry.Key)
		value := e.escapeValue(entry.Value)
		builder.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return []byte(builder.String()), nil
}

// escapeValue escapa valores para formato .env
func (e *DotEnvExporter) escapeValue(value string) string {
	needsQuotes := strings.ContainsAny(value, " \t\n\"'#$\\")

	if needsQuotes {
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}

	return value
}
