package elasticache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"go.uber.org/zap"
)

type VersionProvider struct {
	api aws.ElasticacheApi
	log *zap.Logger
}

func NewProvider(log *zap.Logger, api aws.ElasticacheApi) (*VersionProvider, error) {
	return &VersionProvider{
		api: api,
		log: log,
	}, nil
}

func (vp *VersionProvider) LoadCandidates(soft *software.Software, engine string, filter filters.Filter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := vp.api.DescribeCacheEngineVersions(ctx, &elasticache.DescribeCacheEngineVersionsInput{})
	if err != nil {
		return err
	}
	engineVersions := []types.CacheEngineVersion{}
	for _, v := range res.CacheEngineVersions {
		if *v.Engine == engine {
			engineVersions = append(engineVersions, v)
		}
	}

	for _, version := range engineVersions {
		candidate := software.Version{Version: *version.EngineVersion}
		keep := filter(soft.Version, candidate)
		if keep {
			soft.VersionCandidates = append(soft.VersionCandidates, software.Version{Version: *version.EngineVersion})
		}
	}
	return nil
}
