package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		configFilePath  string
		expectedSuccess bool
	}{
		{
			configFilePath:  "./_testdata/config-success.yml",
			expectedSuccess: true,
		},
		{
			configFilePath:  "./_testdata/config-fail-no-sources.yml",
			expectedSuccess: false,
		},
	}
	for i, tc := range testCases {
		_, err := Load(tc.configFilePath)
		if err != nil && tc.expectedSuccess {
			t.Errorf("Case %d, expected success but got error %s", i+1, err)
		}
		if err == nil && !tc.expectedSuccess {
			t.Errorf("Case %d, expected error but got success", i+1)
		}
	}
}
