package amazonaws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func NewAwsConfig() *aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithDefaultRegion("us-east-1"))
	if err != nil {
		panic(err)
	}
	return &cfg
}

func NewS3Client(cfg *aws.Config) *s3.Client {
	return s3.NewFromConfig(*cfg)
}

func NewDynamodbClient(cfg *aws.Config) *dynamodb.Client {
	return dynamodb.NewFromConfig(*cfg)
}

func NewSsmClient(cfg *aws.Config) *ssm.Client {
	return ssm.NewFromConfig(*cfg)
}

func NewSnsClient(cfg *aws.Config) *sns.Client {
	return sns.NewFromConfig(*cfg)
}
