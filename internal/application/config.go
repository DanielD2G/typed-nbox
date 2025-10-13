package application

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CredentialsSource string

const (
	// SourceEnvVar carga credenciales desde variable de entorno (desarrollo)
	SourceEnvVar CredentialsSource = "env"
	// SourceSecretsManager carga credenciales desde AWS Secrets Manager (producción)
	SourceSecretsManager CredentialsSource = "secretsmanager"
	// SourceFile carga credenciales desde archivo local (desarrollo/testing)
	SourceFile CredentialsSource = "file"
)

type CredentialsLoaderConfig struct {
	Source    CredentialsSource
	EnvVarKey string // Nombre de la variable de entorno
	SecretARN string // ARN del secreto en AWS Secrets Manager
	FilePath  string // Ruta al archivo de credenciales
}

type Config struct {
	BucketName             string   `pkl:"bucketName"`
	EntryTableName         string   `pkl:"entryTableName"`
	TrackingEntryTableName string   `pkl:"trackingEntryTableName"`
	BoxTableName           string   `pkl:"boxTableName"`
	RegionName             string   `pkl:"regionName"`
	AccountId              string   `pkl:"accountId"`
	ParameterStoreKeyId    string   `pkl:"parameterStoreKeyId"`
	ParameterShortArn      bool     `pkl:"parameterShortArn"`
	DefaultPrefix          string   `pkl:"defaultPrefix"`
	AllowedPrefixes        []string `pkl:"allowedPrefixes"`
	HmacSecretKey          []byte
	CredentialsLoader      CredentialsLoaderConfig
}

// #nosec G101
const EnvCredentials = "NBOX_BASIC_AUTH_CREDENTIALS"

func NewConfigFromEnv() *Config {
	var prefixes []string

	defaultPrefix := env("NBOX_DEFAULT_PREFIX", "global")

	prefixes = append(prefixes, fmt.Sprintf("%s/", defaultPrefix))

	prefixes = append(
		prefixes,
		strings.Split(env("NBOX_ALLOWED_PREFIXES", "development/,qa/,beta/,staging/,sandbox/,production/"), ",")...,
	)

	// Configurar estrategia de carga de credenciales
	credSource := CredentialsSource(env("NBOX_CREDENTIALS_SOURCE", "env"))
	credConfig := CredentialsLoaderConfig{
		Source:    credSource,
		EnvVarKey: EnvCredentials,
		SecretARN: env("NBOX_CREDENTIALS_SECRET_ARN", ""),
		FilePath:  env("NBOX_CREDENTIALS_FILE", ".credentials.json"),
	}

	return &Config{
		BucketName:             env("NBOX_BUCKET_NAME", "nbox-store"),
		EntryTableName:         env("NBOX_ENTRIES_TABLE_NAME", "nbox-entry-table"),
		TrackingEntryTableName: env("NBOX_TRACKING_ENTRIES_TABLE_NAME", "nbox-tracking-entry-table"),
		BoxTableName:           env("NBOX_BOX_TABLE_NAME", "nbox-box-table"),
		AccountId:              env("ACCOUNT_ID", ""),
		RegionName:             env("AWS_REGION", "us-east-1"),
		ParameterStoreKeyId:    env("NBOX_PARAMETER_STORE_KEY_ID", ""), // KMS KEY ID
		ParameterShortArn:      envBool("NBOX_PARAMETER_STORE_SHORT_ARN"),
		DefaultPrefix:          defaultPrefix,
		AllowedPrefixes:        prefixes,
		HmacSecretKey:          []byte(env("HMAC_SECRET_KEY", "")),
		CredentialsLoader:      credConfig,
	}
}

func env(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func envBool(key string) bool {
	s := env(key, "false")
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return v
}
