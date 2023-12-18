package versions

import "testing"

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
