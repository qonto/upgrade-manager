package filters

import (
	goversion "github.com/hashicorp/go-version"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
)

type SemverVersionsConfig struct {
	RemovePreRelease        bool `yaml:"remove-pre-release"`
	RemoveFirstMajorVersion bool `yaml:"remove-first-major-version"`
}

func SemverVersions(config SemverVersionsConfig) Filter {
	return func(currentVersion software.Version, candidateVersion software.Version) bool {
		current, err := goversion.NewSemver(currentVersion.Version)
		if err != nil {
			// keep invalid semver releases
			return true
		}
		candidate, err := goversion.NewSemver(candidateVersion.Version)
		if err != nil {
			return true
		}
		candidateSegments := candidate.Segments()
		if current.GreaterThan(candidate) {
			return false
		}
		if current.Equal(candidate) {
			return false
		}
		min, patch := candidateSegments[1], candidateSegments[2]
		if config.RemovePreRelease && candidate.Prerelease() != "" {
			return false
		}
		if config.RemoveFirstMajorVersion && min == 0 && patch == 0 {
			return false
		}
		return true
	}
}
