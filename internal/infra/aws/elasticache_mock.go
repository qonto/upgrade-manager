package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
)

type ElasticacheMock struct{}

func (m *ElasticacheMock) DescribeCacheClusters(ctx context.Context, params *elasticache.DescribeCacheClustersInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
	return &elasticache.DescribeCacheClustersOutput{
		CacheClusters: []types.CacheCluster{
			{
				ARN:                aws.String("someRandomArn1"),
				CacheClusterId:     aws.String("redis1-001"),
				Engine:             aws.String("redis"),
				ReplicationGroupId: aws.String("redis1"),
				EngineVersion:      aws.String("1.2.3"),
			},
			{
				ARN:                aws.String("someRandomArn2"),
				CacheClusterId:     aws.String("redis1-002"),
				Engine:             aws.String("redis"),
				ReplicationGroupId: aws.String("redis1"),
				EngineVersion:      aws.String("1.2.3"),
			},
			{
				ARN:                aws.String("someRandomArn3"),
				CacheClusterId:     aws.String("automation-001"),
				Engine:             aws.String("redis"),
				ReplicationGroupId: aws.String("automation"),
				EngineVersion:      aws.String("1.2.3"),
			},
			{
				ARN:                aws.String("someRandomArn4"),
				CacheClusterId:     aws.String("automation-002"),
				Engine:             aws.String("redis"),
				ReplicationGroupId: aws.String("automation"),
				EngineVersion:      aws.String("1.2.3"),
			},
			{
				ARN:                aws.String("someRandomArn5"),
				CacheClusterId:     aws.String("memcached1-001"),
				Engine:             aws.String("memcached"),
				ReplicationGroupId: aws.String("memcached1"),
				EngineVersion:      aws.String("1.2.3"),
			},
			{
				ARN:                aws.String("someRandomArn6"),
				CacheClusterId:     aws.String("memcached1-002"),
				Engine:             aws.String("memcached"),
				ReplicationGroupId: aws.String("memcached1"),
				EngineVersion:      aws.String("1.2.3"),
			},
		},
	}, nil
}

func (m *ElasticacheMock) DescribeCacheEngineVersions(ctx context.Context, params *elasticache.DescribeCacheEngineVersionsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeCacheEngineVersionsOutput, error) {
	return &elasticache.DescribeCacheEngineVersionsOutput{
		CacheEngineVersions: []types.CacheEngineVersion{
			{
				Engine:        aws.String("redis"),
				EngineVersion: aws.String("7.0.1"),
			},
			{
				Engine:        aws.String("redis"),
				EngineVersion: aws.String("7.0.0"),
			},
			{
				Engine:        aws.String("redis"),
				EngineVersion: aws.String("6.2.6"),
			},
			{
				Engine:        aws.String("redis"),
				EngineVersion: aws.String("6.0.0"),
			},
			{
				Engine:        aws.String("memcached"),
				EngineVersion: aws.String("5.0.1"),
			},
			{
				Engine:        aws.String("memcached"),
				EngineVersion: aws.String("5.0.0"),
			},
			{
				Engine:        aws.String("memcached"),
				EngineVersion: aws.String("4.2.4"),
			},
			{
				Engine:        aws.String("memcached"),
				EngineVersion: aws.String("4.0.0"),
			},
		},
	}, nil
}
