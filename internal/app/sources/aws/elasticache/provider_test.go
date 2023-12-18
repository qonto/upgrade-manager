package elasticache

import (
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"go.uber.org/zap"
)

func TestLoad(t *testing.T) {
	mockApi := new(aws.ElasticacheMock)
	testCases := []struct {
		app                           *software.Software
		filters                       filters.Config
		softType                      software.SoftwareType
		expectedVersionCandidateCount int
	}{
		{
			app: &software.Software{
				Version: software.Version{Version: "6.0.0"},
				Type:    RedisElasticacheCluster,
			},
			filters: filters.Config{
				SemverVersions: &filters.SemverVersionsConfig{
					RemovePreRelease:        true,
					RemoveFirstMajorVersion: true,
				},
			},

			expectedVersionCandidateCount: 2,
		},
		{
			app: &software.Software{
				Version: software.Version{Version: "6.0.0"},
				Type:    RedisElasticacheCluster,
			},
			filters: filters.Config{
				SemverVersions: &filters.SemverVersionsConfig{
					RemovePreRelease:        true,
					RemoveFirstMajorVersion: false,
				},
			},
			expectedVersionCandidateCount: 3,
		},
		{
			app: &software.Software{
				Version: software.Version{Version: "4.2.4"},
				Type:    MemcachedElasticacheCluster,
			},
			filters: filters.Config{
				SemverVersions: &filters.SemverVersionsConfig{
					RemovePreRelease:        true,
					RemoveFirstMajorVersion: true,
				},
			},
			expectedVersionCandidateCount: 1,
		},
		{
			app: &software.Software{
				Version: software.Version{Version: "4.2.4"},
				Type:    MemcachedElasticacheCluster,
			},
			filters: filters.Config{
				SemverVersions: &filters.SemverVersionsConfig{
					RemovePreRelease:        true,
					RemoveFirstMajorVersion: false,
				},
			},
			expectedVersionCandidateCount: 2,
		},
		{
			app: &software.Software{
				Version: software.Version{Version: "5.0.1"},
				Type:    MemcachedElasticacheCluster,
			},
			filters: filters.Config{
				SemverVersions: &filters.SemverVersionsConfig{
					RemovePreRelease:        true,
					RemoveFirstMajorVersion: false,
				},
			},
			expectedVersionCandidateCount: 0,
		},
	}
	for idx, tc := range testCases {
		vp, err := NewProvider(zap.NewExample(), mockApi)
		if err != nil {
			t.Error(err)
		}
		var engine string
		switch tc.app.Type {
		case MemcachedElasticacheCluster:
			engine = "memcached"
		case RedisElasticacheCluster:
			engine = "redis"
		}
		filter := filters.Build(tc.filters)
		err = vp.LoadCandidates(tc.app, engine, filter)
		if err != nil {
			t.Error(err)
		}
		if len(tc.app.VersionCandidates) != tc.expectedVersionCandidateCount {
			t.Errorf("Case %d: Wrong version candidate count. Expected: %d, got : %d)", idx+1, tc.expectedVersionCandidateCount, len(tc.app.VersionCandidates))
		}
	}
}
