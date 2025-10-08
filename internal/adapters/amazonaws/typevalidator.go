package amazonaws

import (
	"context"
	"errors"
	"nbox/internal/application"
	"nbox/internal/domain"
	"nbox/internal/domain/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

type TypeValidatorRecord struct {
	Name  string `dynamodbav:"Name"`
	Regex string `dynamodbav:"Regex"`
}

type typeValidatorBackend struct {
	client *dynamodb.Client
	config *application.Config
	logger *zap.Logger
}

func NewTypeValidatorBackend(dynamodb *dynamodb.Client, config *application.Config, logger *zap.Logger) domain.TypeValidatorAdapter {
	return &typeValidatorBackend{
		client: dynamodb,
		config: config,
		logger: logger,
	}
}

// Upsert creates or updates a type validator
func (t *typeValidatorBackend) Upsert(ctx context.Context, validator models.TypeValidator) error {
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
		TableName: aws.String(t.config.TypeValidatorTableName),
		Item:      item,
	})

	if err != nil {
		t.logger.Error("ErrPutItem", zap.Error(err))
		return err
	}

	return nil
}

// Retrieve gets a type validator by name
func (t *typeValidatorBackend) Retrieve(ctx context.Context, name string) (*models.TypeValidator, error) {
	// Check built-in validators first
	if validator, exists := models.BuiltInValidators[name]; exists {
		return &validator, nil
	}

	// Look up in DynamoDB
	nameAttr, _ := attributevalue.Marshal(name)

	resp, err := t.client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:            map[string]types.AttributeValue{"Name": nameAttr},
		TableName:      aws.String(t.config.TypeValidatorTableName),
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
func (t *typeValidatorBackend) List(ctx context.Context) ([]models.TypeValidator, error) {
	validators := make([]models.TypeValidator, 0)

	// Add built-in validators
	for _, validator := range models.BuiltInValidators {
		validators = append(validators, validator)
	}

	// Scan custom validators from DynamoDB
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(t.config.TypeValidatorTableName),
	}

	scanPaginator := dynamodb.NewScanPaginator(t.client, scanInput)
	var response *dynamodb.ScanOutput
	var err error

	for scanPaginator.HasMorePages() {
		response, err = scanPaginator.NextPage(ctx)
		if err != nil {
			t.logger.Error("ErrScanPaginator", zap.Error(err))
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
	}

	return validators, nil
}

// Delete removes a custom type validator
func (t *typeValidatorBackend) Delete(ctx context.Context, name string) error {
	// Check if it's a built-in validator
	if _, isBuiltIn := models.BuiltInValidators[name]; isBuiltIn {
		return errors.New("cannot delete built-in type validator")
	}

	nameAttr, _ := attributevalue.Marshal(name)

	_, err := t.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(t.config.TypeValidatorTableName),
		Key:       map[string]types.AttributeValue{"Name": nameAttr},
	})

	if err != nil {
		t.logger.Error("ErrDeleteItem", zap.Error(err))
		return err
	}

	return nil
}