package gitutils

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type RepoConnectionProvider interface {
	Clone(directory string, repoUrl string, revision string) (*git.Repository, error) // Clone a git repository to the directory specified
}

type RepoConnection struct {
	Url     string
	Auth    transport.AuthMethod
	Private bool
}

// type GitUrlType string

// const SshUrl gitRepoUrl = "sshUrl"
// const HttpUrl gitRepoUrl = "httpUrl"

func (rc *RepoConnection) Clone(directory string, repoUrl string, revision string) (*git.Repository, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return &git.Repository{}, err
	}
	r, err := git.PlainClone(directory, false,
		&git.CloneOptions{
			URL:           repoUrl,
			ReferenceName: plumbing.ReferenceName("refs/heads/" + revision),
			SingleBranch:  true,
			Depth:         1,
			Auth:          rc.Auth,
		})
	return r, err
}

// func IsGitUrlPrefix(u string) bool {
// 	r, _ := regexp.Compile("(?:git|ssh|https?|git@:.+):(.*)")
// 	return r.MatchString(u)
// }
