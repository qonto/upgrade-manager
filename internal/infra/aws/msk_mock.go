package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/stretchr/testify/mock"
)

type MockMSKApi struct {
	mock.Mock
}

func (m *MockMSKApi) GetCompatibleKafkaVersions(ctx context.Context, params *kafka.GetCompatibleKafkaVersionsInput, optFns ...func(*kafka.Options)) (*kafka.GetCompatibleKafkaVersionsOutput, error) {
	args := m.Called()
	return args.Get(0).(*kafka.GetCompatibleKafkaVersionsOutput), nil //nolint
}

func (m *MockMSKApi) ListClustersV2(ctx context.Context, params *kafka.ListClustersV2Input, optFns ...func(*kafka.Options)) (*kafka.ListClustersV2Output, error) {
	args := m.Called()
	return args.Get(0).(*kafka.ListClustersV2Output), nil //nolint
}
