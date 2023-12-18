package filters

import (
	"testing"
	"time"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/stretchr/testify/assert"
)

func TestRecentVersions(t *testing.T) {
	filter := RecentVersions(RecentVersionsConfig{Days: 20})

	cases := []filterTestCase{
		{
			Current:   software.Version{},
			Candidate: software.Version{Version: "1.1.1"},
			Result:    true,
		},
		{
			Current:   software.Version{},
			Candidate: software.Version{Version: "1.1.1", ReleaseDate: time.Now().Add(-24 * 30 * time.Hour)},
			Result:    true,
		},
		{
			Current:   software.Version{},
			Candidate: software.Version{Version: "1.1.1", ReleaseDate: time.Now().Add(-24 * 21 * time.Hour)},
			Result:    true,
		},
		{
			Current:   software.Version{},
			Candidate: software.Version{Version: "1.1.1", ReleaseDate: time.Now().Add(-24 * 19 * time.Hour)},
			Result:    false,
		},
		{
			Current:   software.Version{},
			Candidate: software.Version{Version: "1.1.1", ReleaseDate: time.Now().Add(-24 * 10 * time.Hour)},
			Result:    false,
		},
	}
	for _, c := range cases {
		result := filter(c.Current, c.Candidate)
		assert.Equalf(t, result, c.Result, "wrong result on versions %s/%s", c.Current.Version, c.Candidate.Version)
	}
}
