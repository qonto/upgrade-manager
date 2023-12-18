package filters

import (
	"github.com/qonto/upgrade-manager/internal/app/core/software"
)

type Config struct {
	SemverVersions *SemverVersionsConfig `yaml:"semver-versions"`
	RecentVersions *RecentVersionsConfig `yaml:"recent-versions"`
}

type Filter func(currentVersion software.Version, candidateVersion software.Version) bool

func Build(config Config) Filter {
	filters := []Filter{}
	if config.SemverVersions != nil {
		filters = append(filters, SemverVersions(*config.SemverVersions))
	}
	if config.RecentVersions != nil {
		filters = append(filters, RecentVersions(*config.RecentVersions))
	}
	return func(currentVersion software.Version, candidateVersion software.Version) bool {
		for _, filter := range filters {
			result := filter(currentVersion, candidateVersion)

			if !result {
				return false
			}
		}
		return true
	}
}
