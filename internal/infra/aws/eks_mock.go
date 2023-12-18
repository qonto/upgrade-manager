package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/stretchr/testify/mock"
)

type EksMock struct {
	mock.Mock
}

func (m *EksMock) ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	args := m.Called()
	return args.Get(0).(*eks.ListClustersOutput), nil //nolint
}

func (m *EksMock) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	args := m.Called(*params)
	return args.Get(0).(*eks.DescribeClusterOutput), nil //nolint
}

func (m *EksMock) DescribeAddon(ctx context.Context, params *eks.DescribeAddonInput, optFns ...func(*eks.Options)) (*eks.DescribeAddonOutput, error) {
	args := m.Called(*params)
	return args.Get(0).(*eks.DescribeAddonOutput), nil //nolint
}

func (m *EksMock) DescribeAddonVersions(ctx context.Context, params *eks.DescribeAddonVersionsInput, optFns ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
	args := m.Called()
	return args.Get(0).(*eks.DescribeAddonVersionsOutput), nil //nolint
}

func (m *EksMock) ListAddons(ctx context.Context, params *eks.ListAddonsInput, optFns ...func(*eks.Options)) (*eks.ListAddonsOutput, error) {
	args := m.Called(*params)
	return args.Get(0).(*eks.ListAddonsOutput), nil //nolint
}
