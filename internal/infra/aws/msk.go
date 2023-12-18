package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
)

type MSKApi interface {
	GetCompatibleKafkaVersions(ctx context.Context, params *kafka.GetCompatibleKafkaVersionsInput, optFns ...func(*kafka.Options)) (*kafka.GetCompatibleKafkaVersionsOutput, error)
	ListClustersV2(ctx context.Context, params *kafka.ListClustersV2Input, optFns ...func(*kafka.Options)) (*kafka.ListClustersV2Output, error)
}
