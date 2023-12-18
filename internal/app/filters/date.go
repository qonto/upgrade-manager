package filters

import (
	"time"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
)

type RecentVersionsConfig struct {
	Days uint `yaml:"days"`
}

func RecentVersions(config RecentVersionsConfig) Filter {
	return func(_ software.Version, candidateVersion software.Version) bool {
		if candidateVersion.ReleaseDate.IsZero() {
			return true
		}
		return candidateVersion.ReleaseDate.Add(time.Duration(config.Days) * 24 * time.Hour).Before(time.Now())
	}
}
