package filters

import (
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/stretchr/testify/assert"
)

type filterTestCase struct {
	Current   software.Version
	Candidate software.Version
	Result    bool
}

func TestSemverVersionsIgnorePreRelease(t *testing.T) {
	filter := SemverVersions(SemverVersionsConfig{RemovePreRelease: true})

	cases := []filterTestCase{
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1-rc"},

			Result: false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.0"},

			Result: false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1"},

			Result: true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.0.0"},

			Result: false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "0.0.1"},

			Result: false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "2.0.0"},

			Result: true,
		},
		{
			Current:   software.Version{Version: "1.1.0a"},
			Candidate: software.Version{Version: "2.0.0"},
			Result:    true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "2.0.0a"},
			Result:    true,
		},
	}
	for _, c := range cases {
		result := filter(c.Current, c.Candidate)
		assert.Equalf(t, result, c.Result, "wrong result on versions %s/%s", c.Current.Version, c.Candidate.Version)
	}
}

func TestSemverVersionsIgnoreFirstMajor(t *testing.T) {
	filter := SemverVersions(SemverVersionsConfig{RemoveFirstMajorVersion: true})

	cases := []filterTestCase{
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1-rc"},
			Result:    true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "2.0.1"},
			Result:    true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "2.0.0"},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "3.0.0"},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.1.1"},
			Result:    true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "1.0.0"},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "0.0.1"},
			Result:    false,
		},
		{
			Current:   software.Version{Version: "1.1.0a"},
			Candidate: software.Version{Version: "2.0.0"},
			Result:    true,
		},
		{
			Current:   software.Version{Version: "1.1.0"},
			Candidate: software.Version{Version: "2.0.0a"},
			Result:    true,
		},
	}
	for _, c := range cases {
		result := filter(c.Current, c.Candidate)
		assert.Equalf(t, result, c.Result, "wrong result on versions %s/%s", c.Current.Version, c.Candidate.Version)
	}
}
