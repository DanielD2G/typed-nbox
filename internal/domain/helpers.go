package domain

import "strings"

// ConvertToEnvVarName convierte una key de NBOX a formato de variable de entorno
// example: "production/myapp/database/host" -> "PRODUCTION_MYAPP_DATABASE_HOST"
func ConvertToEnvVarName(key string) string {
	// Reemplazar caracteres no permitidos
	envVar := strings.ToUpper(key)
	envVar = strings.ReplaceAll(envVar, "/", "_")
	envVar = strings.ReplaceAll(envVar, "-", "_")
	envVar = strings.ReplaceAll(envVar, ".", "_")
	envVar = strings.ReplaceAll(envVar, " ", "_")

	// Remover caracteres especiales
	var cleaned strings.Builder
	for _, r := range envVar {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String()
}
