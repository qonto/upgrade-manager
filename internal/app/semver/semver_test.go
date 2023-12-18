package semver

import (
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
)

func TestSortSoftwareVersions(t *testing.T) {
	testCases := []struct {
		versions []software.Version
		expected string
	}{
		{
			versions: []software.Version{
				{
					Version: "7.0.0",
				},
				{
					Version: "6.0.0",
				},
				{
					Version: "5.0.0",
				},
			},
			expected: "7.0.0",
		},
		{
			versions: []software.Version{
				{
					Version: "5.0.0",
				},
				{
					Version: "6.0.0",
				},
				{
					Version: "7.0.0",
				},
			},
			expected: "7.0.0",
		},
	}
	for idx, testCase := range testCases {
		Sort(testCase.versions)
		if testCase.versions[0].Version != testCase.expected {
			t.Errorf("Case %d, wrong first element in sorted slice. Expected %s, got: %s", idx+1, testCase.expected, testCase.versions[0].Version)
		}
	}
}

func TestExtractFromString(t *testing.T) {
	versions := software.Versions{
		{Version: "1.2.3"},
		{Version: "1.2.3-eksbuild.v2"},
		{Version: "1.2.3-alpha"},
	}
	for _, v := range versions {
		semVer, err := ExtractFromString(v.Version)
		if semVer != "1.2.3" || err != nil {
			t.Errorf("error while parsing semantic version")
		}
	}
}
