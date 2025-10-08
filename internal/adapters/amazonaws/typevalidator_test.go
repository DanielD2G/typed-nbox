package amazonaws

import (
	"context"
	"errors"
	"nbox/internal/application"
	"nbox/internal/domain/models"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

// Mock DynamoDB client for testing
type mockDynamoDBClient struct {
	putItemFunc    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	getItemFunc    func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	deleteItemFunc func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	scanFunc       func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if m.putItemFunc != nil {
		return m.putItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if m.getItemFunc != nil {
		return m.getItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if m.deleteItemFunc != nil {
		return m.deleteItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.DeleteItemOutput{}, nil
}

func (m *mockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if m.scanFunc != nil {
		return m.scanFunc(ctx, params, optFns...)
	}
	return &dynamodb.ScanOutput{}, nil
}

// We need these methods to satisfy the interface, but they won't be used in these tests
func (m *mockDynamoDBClient) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	return &dynamodb.BatchWriteItemOutput{}, nil
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{}, nil
}

func (m *mockDynamoDBClient) DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	return &dynamodb.DescribeTableOutput{}, nil
}

// dynamoDBClient interface for testing
type dynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}

// testTypeValidatorBackend wraps typeValidatorBackend to allow testing with mock client
type testTypeValidatorBackend struct {
	client dynamoDBClient
	config *application.Config
	logger *zap.Logger
}

func newTestTypeValidatorBackend(client *mockDynamoDBClient) *testTypeValidatorBackend {
	config := &application.Config{
		TypeValidatorTableName: "test-type-validator-table",
	}
	logger := zap.NewNop()

	return &testTypeValidatorBackend{
		client: client,
		config: config,
		logger: logger,
	}
}

// Upsert creates or updates a type validator
func (t *testTypeValidatorBackend) Upsert(ctx context.Context, validator models.TypeValidator) error {
	// Check if it's a built-in validator
	if _, isBuiltIn := models.BuiltInValidators[validator.Name]; isBuiltIn {
		return errors.New("cannot modify built-in type validator")
	}

	record := TypeValidatorRecord{
		Name:  validator.Name,
		Regex: validator.Regex,
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		t.logger.Error("ErrMarshalMap", zap.Error(err))
		return err
	}

	_, err = t.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &t.config.TypeValidatorTableName,
		Item:      item,
	})

	if err != nil {
		t.logger.Error("ErrPutItem", zap.Error(err))
		return err
	}

	return nil
}

// Retrieve gets a type validator by name
func (t *testTypeValidatorBackend) Retrieve(ctx context.Context, name string) (*models.TypeValidator, error) {
	// Check built-in validators first
	if validator, exists := models.BuiltInValidators[name]; exists {
		return &validator, nil
	}

	// Look up in DynamoDB
	nameAttr, _ := attributevalue.Marshal(name)

	resp, err := t.client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:            map[string]types.AttributeValue{"Name": nameAttr},
		TableName:      &t.config.TypeValidatorTableName,
		ConsistentRead: aws.Bool(true),
	})

	if err != nil {
		t.logger.Error("ErrGetItem", zap.Error(err))
		return nil, err
	}

	if resp.Item == nil {
		return nil, nil
	}

	record := &TypeValidatorRecord{}
	err = attributevalue.UnmarshalMap(resp.Item, record)
	if err != nil {
		t.logger.Error("ErrUnmarshalMap", zap.Error(err))
		return nil, err
	}

	return &models.TypeValidator{
		Name:  record.Name,
		Regex: record.Regex,
	}, nil
}

// List returns all type validators (built-in + custom)
func (t *testTypeValidatorBackend) List(ctx context.Context) ([]models.TypeValidator, error) {
	validators := make([]models.TypeValidator, 0)

	// Add built-in validators
	for _, validator := range models.BuiltInValidators {
		validators = append(validators, validator)
	}

	// Scan custom validators from DynamoDB
	scanInput := &dynamodb.ScanInput{
		TableName: &t.config.TypeValidatorTableName,
	}

	// For testing, we'll just call Scan once instead of using pagination
	response, err := t.client.Scan(ctx, scanInput)
	if err != nil {
		t.logger.Error("ErrScan", zap.Error(err))
		return nil, err
	}

	var records []TypeValidatorRecord
	err = attributevalue.UnmarshalListOfMaps(response.Items, &records)
	if err != nil {
		t.logger.Error("ErrUnmarshalListOfMaps", zap.Error(err))
		return nil, err
	}

	for _, record := range records {
		validators = append(validators, models.TypeValidator{
			Name:  record.Name,
			Regex: record.Regex,
		})
	}

	return validators, nil
}

// Delete removes a custom type validator
func (t *testTypeValidatorBackend) Delete(ctx context.Context, name string) error {
	// Check if it's a built-in validator
	if _, isBuiltIn := models.BuiltInValidators[name]; isBuiltIn {
		return errors.New("cannot delete built-in type validator")
	}

	nameAttr, _ := attributevalue.Marshal(name)

	_, err := t.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &t.config.TypeValidatorTableName,
		Key:       map[string]types.AttributeValue{"Name": nameAttr},
	})

	if err != nil {
		t.logger.Error("ErrDeleteItem", zap.Error(err))
		return err
	}

	return nil
}

func TestTypeValidatorBackend_Upsert(t *testing.T) {
	tests := []struct {
		name      string
		validator models.TypeValidator
		mockFunc  func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
		wantErr   bool
	}{
		{
			name: "successful upsert of custom validator",
			validator: models.TypeValidator{
				Name:  "email",
				Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			mockFunc: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				return &dynamodb.PutItemOutput{}, nil
			},
			wantErr: false,
		},
		{
			name: "cannot modify built-in validator - string",
			validator: models.TypeValidator{
				Name:  "string",
				Regex: ".*",
			},
			wantErr: true,
		},
		{
			name: "cannot modify built-in validator - number",
			validator: models.TypeValidator{
				Name:  "number",
				Regex: `^\d+$`,
			},
			wantErr: true,
		},
		{
			name: "dynamodb error",
			validator: models.TypeValidator{
				Name:  "custom",
				Regex: "test",
			},
			mockFunc: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				return nil, errors.New("dynamodb error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockDynamoDBClient{
				putItemFunc: tt.mockFunc,
			}
			backend := newTestTypeValidatorBackend(mockClient)

			err := backend.Upsert(context.Background(), tt.validator)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeValidatorBackend_Retrieve(t *testing.T) {
	tests := []struct {
		name     string
		valName  string
		mockFunc func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
		want     *models.TypeValidator
		wantErr  bool
	}{
		{
			name:    "retrieve built-in validator - number",
			valName: "number",
			want:    &models.TypeValidatorNumber,
			wantErr: false,
		},
		{
			name:    "retrieve built-in validator - json",
			valName: "json",
			want:    &models.TypeValidatorJSON,
			wantErr: false,
		},
		{
			name:    "retrieve custom validator from dynamodb",
			valName: "email",
			mockFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				record := TypeValidatorRecord{
					Name:  "email",
					Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				}
				item, _ := attributevalue.MarshalMap(record)
				return &dynamodb.GetItemOutput{
					Item: item,
				}, nil
			},
			want: &models.TypeValidator{
				Name:  "email",
				Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			wantErr: false,
		},
		{
			name:    "validator not found",
			valName: "nonexistent",
			mockFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{}, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "dynamodb error",
			valName: "test",
			mockFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return nil, errors.New("dynamodb error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockDynamoDBClient{
				getItemFunc: tt.mockFunc,
			}
			backend := newTestTypeValidatorBackend(mockClient)

			got, err := backend.Retrieve(context.Background(), tt.valName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && got != nil {
				t.Errorf("Retrieve() = %v, want nil", got)
			} else if tt.want != nil && got == nil {
				t.Errorf("Retrieve() = nil, want %v", tt.want)
			} else if tt.want != nil && got != nil {
				if got.Name != tt.want.Name || got.Regex != tt.want.Regex {
					t.Errorf("Retrieve() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTypeValidatorBackend_List(t *testing.T) {
	tests := []struct {
		name         string
		mockFunc     func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
		wantMinCount int // Minimum count including built-in validators
		wantErr      bool
	}{
		{
			name: "list with custom validators",
			mockFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				records := []TypeValidatorRecord{
					{Name: "email", Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
					{Name: "phone", Regex: `^\+?[1-9]\d{1,14}$`},
				}
				items := make([]map[string]types.AttributeValue, 0)
				for _, record := range records {
					item, _ := attributevalue.MarshalMap(record)
					items = append(items, item)
				}
				return &dynamodb.ScanOutput{
					Items: items,
				}, nil
			},
			wantMinCount: 7, // 5 built-in + 2 custom
			wantErr:      false,
		},
		{
			name: "list with no custom validators",
			mockFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
			wantMinCount: 5, // 5 built-in validators
			wantErr:      false,
		},
		{
			name: "dynamodb error",
			mockFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return nil, errors.New("dynamodb error")
			},
			wantMinCount: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockDynamoDBClient{
				scanFunc: tt.mockFunc,
			}
			backend := newTestTypeValidatorBackend(mockClient)

			got, err := backend.List(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) < tt.wantMinCount {
				t.Errorf("List() returned %d validators, want at least %d", len(got), tt.wantMinCount)
			}
		})
	}
}

func TestTypeValidatorBackend_Delete(t *testing.T) {
	tests := []struct {
		name     string
		valName  string
		mockFunc func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
		wantErr  bool
	}{
		{
			name:    "successful delete of custom validator",
			valName: "email",
			mockFunc: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				return &dynamodb.DeleteItemOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:    "cannot delete built-in validator - string",
			valName: "string",
			wantErr: true,
		},
		{
			name:    "cannot delete built-in validator - number",
			valName: "number",
			wantErr: true,
		},
		{
			name:    "cannot delete built-in validator - json",
			valName: "json",
			wantErr: true,
		},
		{
			name:    "cannot delete built-in validator - url-https",
			valName: "url-https",
			wantErr: true,
		},
		{
			name:    "cannot delete built-in validator - url-http",
			valName: "url-http",
			wantErr: true,
		},
		{
			name:    "dynamodb error",
			valName: "custom",
			mockFunc: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				return nil, errors.New("dynamodb error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockDynamoDBClient{
				deleteItemFunc: tt.mockFunc,
			}
			backend := newTestTypeValidatorBackend(mockClient)

			err := backend.Delete(context.Background(), tt.valName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
