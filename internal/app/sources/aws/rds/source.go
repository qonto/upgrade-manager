package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"go.uber.org/zap"
)

type Source struct {
	api                     aws.RDSApi
	log                     *zap.Logger
	cfg                     *Config
	filter                  filters.Filter
	engineToSoftTypeMapping map[string]software.SoftwareType
}

const (
	RdsPostgreSQL       software.SoftwareType = "rds-postgresql"
	RdsAuroraPostgreSQL software.SoftwareType = "rds-aurora-postgresql"
	RdsAuroraMySQL      software.SoftwareType = "rds-aurora-mysql"
	RdsMariaDB          software.SoftwareType = "rds-mariadb"
	RdsDocDB            software.SoftwareType = "rds-docdb"
	RdsMySQL            software.SoftwareType = "rds-mysql"
	RdsOracle           software.SoftwareType = "rds-oracle-ee"
	RdsNeptune          software.SoftwareType = "rds-neptune"
	DefaultTimeout      time.Duration         = time.Second * 15
)

func (s *Source) Name() string {
	return "RDS"
}

func NewSource(api aws.RDSApi, log *zap.Logger, cfg *Config) (*Source, error) {
	mapping := map[string]software.SoftwareType{
		"aurora-mysql":      RdsAuroraMySQL,
		"aurora-postgresql": RdsAuroraPostgreSQL,
		"docdb":             RdsDocDB,
		"mariadb":           RdsMariaDB,
		"mysql":             RdsMySQL,
		"neptune":           RdsNeptune,
		"oracle-ee":         RdsOracle,
		"postgres":          RdsPostgreSQL,
	}
	// Current implementation of filters requires this map to be non-nil to filter old versions
	// so we set RemovePreRelease to true to filter out old versions anyway.
	// NOTE: this is slightly confusing and should probably be refactored later on
	cfg.Filters = filters.Config{
		SemverVersions: &filters.SemverVersionsConfig{
			RemovePreRelease: true,
		},
	}
	filter := filters.Build(cfg.Filters)
	return &Source{
		log:                     log,
		api:                     api,
		cfg:                     cfg,
		filter:                  filter,
		engineToSoftTypeMapping: mapping,
	}, nil
}

func (s *Source) Load() ([]*software.Software, error) {
	timeout, err := time.ParseDuration(s.cfg.RequestTimeout)
	if err != nil || s.cfg.RequestTimeout == "" {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var softwares []*software.Software
	pgEngineName := "postgres"
	defaultEngineVersions, err := s.api.DescribeDBEngineVersions(ctx, &rds.DescribeDBEngineVersionsInput{})
	if err != nil {
		return nil, err
	}
	// The below second api call is required. DescribeDBEngineVersions does not return postgres versions by default (only aurora-postgresql)
	pgEngineVersions, err := s.api.DescribeDBEngineVersions(ctx, &rds.DescribeDBEngineVersionsInput{Engine: &pgEngineName})
	if err != nil {
		return nil, err
	}
	defaultEngineVersions.DBEngineVersions = append(defaultEngineVersions.DBEngineVersions, pgEngineVersions.DBEngineVersions...)
	versionRegistry := make(map[string][]software.Version)
	for _, ev := range defaultEngineVersions.DBEngineVersions {
		versionRegistry[*ev.Engine] = append(versionRegistry[*ev.Engine], software.Version{Version: *ev.EngineVersion})
	}
	// Get DB instances info
	res, err := s.api.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, err
	}
	for _, cluster := range res.DBInstances {
		// if aggregation level is clsuter, add the instance only if it is marked as a replica with a source db
		if s.cfg.AggregationLevel == "cluster" {
			if cluster.ReadReplicaSourceDBInstanceIdentifier != nil {
				continue
			}
		}
		// Create software for the DB instance
		cv := software.Version{Version: *cluster.EngineVersion}
		candidates := versionRegistry[*cluster.Engine]
		var versionCandidates []software.Version
		for _, v := range candidates {
			if keep := s.filter(cv, v); keep {
				versionCandidates = append(versionCandidates, v)
			}
		}
		softwares = append(softwares, &software.Software{
			Name:              *cluster.DBInstanceIdentifier,
			Calculator:        software.SemverCalculator,
			Type:              s.engineToSoftTypeMapping[*cluster.Engine],
			Version:           software.Version{Version: *cluster.EngineVersion},
			VersionCandidates: versionCandidates,
		})
		s.log.Info(fmt.Sprintf("Tracking software %s, of type %s", *cluster.DBInstanceIdentifier, s.engineToSoftTypeMapping[*cluster.Engine]))
	}

	return softwares, nil
}
