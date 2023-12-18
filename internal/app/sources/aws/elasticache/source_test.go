package elasticache

import (
	"log/slog"
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
)

func TestSourceLoad(t *testing.T) {
	mockApi := new(aws.ElasticacheMock)
	testCases := []struct {
		cfg                   *Config
		expectedSoftwareCount int
	}{
		{
			cfg: &Config{
				Enabled: true,
				Filters: filters.Config{
					SemverVersions: &filters.SemverVersionsConfig{
						RemovePreRelease:        true,
						RemoveFirstMajorVersion: true,
					},
				},
			},
			expectedSoftwareCount: 3, // REMINDER: per-cluster ID deduplication
		},
	}
	for idx, tc := range testCases {
		source, err := NewSource(mockApi, slog.Default(), tc.cfg)
		if err != nil {
			t.Error(err)
		}
		softwares, err := source.Load()
		if err != nil {
			t.Error(err)
		}
		if len(softwares) != tc.expectedSoftwareCount {
			t.Errorf("Case %d: Wrong software count. Expected: %d, got : %d", idx+1, tc.expectedSoftwareCount, len(softwares))
		}
	}
}
