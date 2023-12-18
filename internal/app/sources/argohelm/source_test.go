package argohelm

import (
	"log/slog"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/qonto/upgrade-manager/internal/app/sources/utils/gitutils"
)

func TestMatchGitRepoConnection(t *testing.T) {
	s := Source{log: slog.Default()}
	s.gitRepoConnections = []*gitutils.RepoConnection{
		{
			Url: "https://github.foo.com/devops/kubernetes-resources/exactmatch.git",
		},
		{
			Url: "git@github.foo.com:devops/kubernetes-resources/exactmatch.git",
		},
		{
			Url: "https://prefix2/",
		},
		{
			Url: "https://prefix3.com",
		},
		{
			Url: "https://prefix1/",
		},
	}
	testCases := []struct {
		url                string
		expectedMatchCount int
	}{
		{
			url:                "https://github.foo.com/devops/kubernetes-resources/exactmatch.git",
			expectedMatchCount: 2, // exact match + prefix for host in https mode
		},
		{
			url:                "git@github.foo.com:devops/kubernetes-resources/exactmatch.git",
			expectedMatchCount: 2, // exact match + prefix for host in https mode
		},
		{
			url:                "https://github.foo.com/devops/kubernetes-resources/onematch.git",
			expectedMatchCount: 1,
		},
		{
			url:                "git@github.foo.com:devops/kubernetes-resources/onematch.git",
			expectedMatchCount: 1,
		},
		{
			url:                "git@github.foo.com:devoopsie/kubernetes-resources/nomatch.git",
			expectedMatchCount: 0,
		},
	}
	for i, tc := range testCases {
		conn, err := s.matchGitRepoConnection(tc.url)
		if err != nil {
			t.Error(err)
		}
		if tc.expectedMatchCount == 0 {
			pass, ok := conn.Auth.(*http.BasicAuth)
			if !ok {
				t.Fatalf("Case %d: Expected 0 match and therefore a default empty basic auth. But was not basic auth", i+1)
			}
			if pass.Password != "" || pass.Username != "" {
				t.Fatalf("Case %d: Expected 0 match and therefore a default empty basic auth. Password : %s, Username: %s", i+1, pass.Password, pass.Username)
			}
		}
	}
}
