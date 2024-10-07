package versions

import (
	"log/slog"
	"testing"

	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/stretchr/testify/assert"
)

func TestGetRepoBackendType(t *testing.T) {
	testCases := map[string]RepoBackendType{
		"https://github.foo.com/repo.git":                     GitRepo,
		"git@github.foo.com:bar/repo.git":                     GitRepo,
		"https://github.com/repo.git":                         GitRepo,
		"https://bitnami-labs.github.io/sealed-secrets":       HelmRepo,
		"https://helm.traefik.io/traefik/traefik-17.0.2.tgz":  HelmRepo,
		"https://helm.traefik.io/traefik":                     HelmRepo,
		"oci://registry-1.docker.io/bitnamicharts/some-chart": HelmRepo,
		"s3://s3-based-repositry":                             S3HelmRepo,
		"s3://s3-based-chart-archive.tgz":                     S3HelmRepo,
	}
	for url, expected := range testCases {
		result, err := getRepoBackendType(url)
		if err != nil {
			t.Errorf("Failed retrieving repo backend type: %s", err)
		}
		if result != expected {
			t.Errorf("Got wrong backend type for %s, got: %s expected: %s", url, result, expected)
		}
	}
}

func TestBuildRepoBackend(t *testing.T) {
	logger := slog.Default()
	mockS3Api := new(aws.S3Mock)

	tests := []struct {
		name         string
		repoAliases  map[string]string
		repoURL      string
		chartName    string
		expectedType interface{}
		expectError  bool
	}{
		{
			name:         "HelmRepo",
			repoAliases:  nil,
			repoURL:      "https://charts.helm.sh/stable",
			chartName:    "mysql",
			expectedType: &HelmRepoBackend{},
			expectError:  false,
		},
		{
			name:         "S3HelmRepo",
			repoAliases:  nil,
			repoURL:      "s3://my-bucket/charts",
			chartName:    "my-chart",
			expectedType: &S3HelmRepoBackend{},
			expectError:  false,
		},
		{
			name:         "GitRepo",
			repoAliases:  nil,
			repoURL:      "https://github.com/user/repo.git",
			chartName:    "my-chart",
			expectedType: nil,
			expectError:  false,
		},
		{
			name:         "Invalid URL",
			repoAliases:  nil,
			repoURL:      "invalid://url",
			chartName:    "my-chart",
			expectedType: nil,
			expectError:  true,
		},
		{
			name:         "With Repo Alias",
			repoAliases:  map[string]string{"@alias": "https://charts.helm.sh/stable"},
			repoURL:      "@alias",
			chartName:    "mysql",
			expectedType: &HelmRepoBackend{},
			expectError:  false,
		},
		{
			name:         "With Repo Alias targeting a S3 bucket",
			repoAliases:  map[string]string{"alias": "s3://my-bucket/charts"},
			repoURL:      "alias",
			chartName:    "my-chart",
			expectedType: &S3HelmRepoBackend{},
			expectError:  false,
		},
		{
			name:         "With Repo Alias targeting a Git repo",
			repoAliases:  map[string]string{"@alias": "https://github.com/user/repo.git"},
			repoURL:      "@alias",
			chartName:    "my-chart",
			expectedType: nil,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := buildRepoBackend(tt.repoAliases, tt.repoURL, tt.chartName, logger, mockS3Api)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedType == nil {
				assert.Nil(t, backend)
			} else {
				assert.IsType(t, tt.expectedType, backend)
			}
		})
	}
}
