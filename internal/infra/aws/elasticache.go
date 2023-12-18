package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

type ElasticacheApi interface {
	DescribeCacheClusters(ctx context.Context, params *elasticache.DescribeCacheClustersInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error)
	DescribeCacheEngineVersions(ctx context.Context, params *elasticache.DescribeCacheEngineVersionsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeCacheEngineVersionsOutput, error)
}
