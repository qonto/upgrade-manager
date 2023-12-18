package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/stretchr/testify/mock"
)

type MockRDSApi struct {
	mock.Mock
}

func (m *MockRDSApi) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	args := m.Called()
	return args.Get(0).(*rds.DescribeDBInstancesOutput), nil //nolint
}

func (m *MockRDSApi) DescribeDBEngineVersions(ctx context.Context, params *rds.DescribeDBEngineVersionsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBEngineVersionsOutput, error) {
	args := m.Called()
	return args.Get(0).(*rds.DescribeDBEngineVersionsOutput), nil //nolint
}
