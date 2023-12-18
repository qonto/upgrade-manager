package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type RDSApi interface {
	DescribeDBEngineVersions(ctx context.Context, params *rds.DescribeDBEngineVersionsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBEngineVersionsOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}
