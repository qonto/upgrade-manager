package elasticache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
)

type Source struct {
	log    *slog.Logger
	api    aws.ElasticacheApi
	cfg    *Config
	vp     *VersionProvider
	filter filters.Filter
}

const (
	RedisElasticacheCluster     software.SoftwareType = "elasticache-redis"
	MemcachedElasticacheCluster software.SoftwareType = "elasticache-memcached"
	DefaultTimeout              time.Duration         = time.Second * 15
)

func (s *Source) Name() string {
	return "elasticache"
}

func NewSource(api aws.ElasticacheApi, log *slog.Logger, cfg *Config) (*Source, error) {
	// Current implementation of filters requires this map to be non-nil to filter old versions
	// so we set RemovePreRelease to true to filter out old versions anyway.
	// NOTE: this is slightly confusing and should probably be refactored later on
	if cfg.Filters.SemverVersions == nil {
		cfg.Filters = filters.Config{
			SemverVersions: &filters.SemverVersionsConfig{
				RemovePreRelease: true,
			},
		}
	}
	chartFilter := filters.Build(cfg.Filters)
	vp, err := NewProvider(log, api)
	if err != nil {
		log.Error("Failed to build elasticache version provider")
		return &Source{}, err
	}
	return &Source{
		log:    log,
		api:    api,
		cfg:    cfg,
		vp:     vp,
		filter: chartFilter,
	}, nil
}

func (s *Source) Load() ([]*software.Software, error) {
	softwares := []*software.Software{}
	timeout, err := time.ParseDuration(s.cfg.RequestTimeout)
	if err != nil || s.cfg.RequestTimeout == "" {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := s.api.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		return nil, err
	}
	processedReplicationGroupId := []string{}

	for _, cluster := range res.CacheClusters {
		clusterProcessed := false
		for _, id := range processedReplicationGroupId {
			// for some reason DescribeCacheClusters return a list of node, so to deduplicate nodes inside a cluster,
			// we check if if the replication group was already
			if *cluster.ReplicationGroupId == id {
				clusterProcessed = true
			}
		}
		if !clusterProcessed {
			processedReplicationGroupId = append(processedReplicationGroupId, *cluster.ReplicationGroupId)

			var softType software.SoftwareType
			switch *cluster.Engine {
			case "redis":
				softType = RedisElasticacheCluster
			case "memcached":
				softType = MemcachedElasticacheCluster
			default:
				s.log.Error(fmt.Sprintf("unknown elasticache cluster type %s, skipping...", *cluster.Engine))
				continue
			}
			soft := &software.Software{
				Name: *cluster.ReplicationGroupId,
				Type: softType,
				Version: software.Version{
					Version: *cluster.EngineVersion,
				},
			}
			s.log.Info(fmt.Sprintf("Tracking software %s of type %s", *cluster.ReplicationGroupId, softType))

			err = s.vp.LoadCandidates(soft, *cluster.Engine, s.filter)
			if err != nil {
				s.log.Warn(fmt.Sprintf("Fail to retrieve versions for software %s: %s", soft.Name, err.Error()))
			}
			softwares = append(softwares, soft)
		}
	}
	return softwares, nil
}
