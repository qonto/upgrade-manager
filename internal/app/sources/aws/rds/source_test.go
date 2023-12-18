package rds

import (
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	dbInfo := []struct {
		dbId      string
		dbEngine  string
		dbVersion string
	}{{"pg1", "postgres", "13.3"}, {"pg1-replica", "postgres", "13.3"}, {"pg2", "mysql", "8.0.23"}}
	dbVersions := map[string][]string{
		"postgres": {"15.2", "14.5", "11.1"},
		"mysql":    {"8.0.28", "8.0.27", "5.7.11"},
	}
	api := new(aws.MockRDSApi)
	api.On("DescribeDBInstances").Return(&rds.DescribeDBInstancesOutput{
		DBInstances: []types.DBInstance{
			{DBInstanceIdentifier: &dbInfo[0].dbId, Engine: &dbInfo[0].dbEngine, EngineVersion: &dbInfo[0].dbVersion},
			{DBInstanceIdentifier: &dbInfo[1].dbId, Engine: &dbInfo[1].dbEngine, EngineVersion: &dbInfo[0].dbVersion, ReadReplicaSourceDBInstanceIdentifier: &dbInfo[0].dbId},
			{DBInstanceIdentifier: &dbInfo[2].dbId, Engine: &dbInfo[2].dbEngine, EngineVersion: &dbInfo[0].dbVersion},
		},
	})
	api.On("DescribeDBEngineVersions").Return(&rds.DescribeDBEngineVersionsOutput{
		DBEngineVersions: []types.DBEngineVersion{
			{Engine: &dbInfo[0].dbEngine, EngineVersion: &dbVersions[dbInfo[0].dbEngine][0]},
			{Engine: &dbInfo[0].dbEngine, EngineVersion: &dbVersions[dbInfo[0].dbEngine][1]},
			{Engine: &dbInfo[0].dbEngine, EngineVersion: &dbVersions[dbInfo[0].dbEngine][2]},
			{Engine: &dbInfo[2].dbEngine, EngineVersion: &dbVersions[dbInfo[2].dbEngine][0]},
			{Engine: &dbInfo[2].dbEngine, EngineVersion: &dbVersions[dbInfo[2].dbEngine][1]},
			{Engine: &dbInfo[2].dbEngine, EngineVersion: &dbVersions[dbInfo[2].dbEngine][2]},
		},
	},
	)

	tcases := []struct {
		cfg                   *Config
		expectedSoftwareCount int
	}{
		{
			cfg: &Config{
				AggregationLevel: "cluster",
			},
			expectedSoftwareCount: 2,
		},
		{
			cfg: &Config{
				AggregationLevel: "instance",
			},
			expectedSoftwareCount: 3,
		},
	}
	for _, tc := range tcases {
		src, err := NewSource(api, slog.Default(), tc.cfg)
		if err != nil {
			t.Error(err)
		}
		softwares, err := src.Load()
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, tc.expectedSoftwareCount, len(softwares))
	}
}
