package filters

import (
	"testing"
	"time"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	config := Config{
		SemverVersions: &SemverVersionsConfig{
			RemovePreRelease:        true,
			RemoveFirstMajorVersion: true,
		},
		RecentVersions: &RecentVersionsConfig{
			Days: 20,
		},
	}
	filter := Build(config)

	cases := []filterTestCase{
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1-rc", ReleaseDate: time.Now().Add(-24 * 30 * time.Hour)},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1-rc", ReleaseDate: time.Now().Add(-24 * 30 * time.Hour)},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1", ReleaseDate: time.Now().Add(-24 * 30 * time.Hour)},
			Result:    true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1", ReleaseDate: time.Now().Add(-24 * 18 * time.Hour)},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "3.0.0-beta2"},
			Result:    false,
		},
	}

	for _, c := range cases {
		result := filter(c.Current, c.Candidate)
		assert.Equalf(t, result, c.Result, "wrong result on versions %s/%s", c.Current.Version, c.Candidate.Version)
	}
}
