package versions

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"time"

	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type RepoBackend interface {
	getIndexFile() (*repo.IndexFile, error)
	getFile(ctx context.Context, url *url.URL) (io.ReadCloser, error)
}

type RepoBackendType string

// Chart reprensented by a Chart.yaml hosted on a remote git repository
const (
	GitRepo RepoBackendType = "gitRepo"

	// Chart Hosted on Chart repository with proper index.yaml
	HelmRepo RepoBackendType = "helmRepo"

	// Chart Hosted on S3 Bucket with proper index.yaml
	S3HelmRepo RepoBackendType = "s3helmRepo"

	// Chart reprensented by a Chart.yaml hosted on the local file-system
	LocalRepo RepoBackendType = "localRepo"

	// When a chart.yaml references a dependency with a file//path/to/chart/directory/ like expression, the file
	// is therefore "local" to the remote git repo. This is not referencing a local git repo on your local desktop
	GitRepoLocal RepoBackendType = "gitRepoLocal"

	// Default timeout duration when calling external helm repositories
	DefaultTimeout time.Duration = time.Second * 15
)

func getRepoBackendType(repoUrl string) (RepoBackendType, error) {
	o := regexp.MustCompile("(?:git|ssh|https?).*.git$")
	u, err := url.Parse(repoUrl)

	if o.MatchString(repoUrl) {
		return GitRepo, nil
	}
	if err != nil {
		r := regexp.MustCompile("(?:git|ssh|ociZ|https?|git@:.+):(.*)")

		if r.MatchString(repoUrl) {
			return GitRepo, nil
		}
		return "", err
	}
	switch u.Scheme {
	case "s3":
		return S3HelmRepo, nil
	case "https", "oci":
		// determine if GitRepo or HelmRepo
		return HelmRepo, nil
	case "file":
		return GitRepoLocal, nil
	}
	return "", fmt.Errorf("could not determine RepoBackendType for url %s", repoUrl)
}

func buildRepoBackend(repoURL string, chartName string, log *slog.Logger, s3Api aws.S3Api) (RepoBackend, error) {
	repoType, err := getRepoBackendType(repoURL)
	if err != nil {
		return nil, err
	}
	var repoBackend RepoBackend
	switch repoType {
	case HelmRepo:
		log.Debug(fmt.Sprintf("Configuring HelmRepoBackend for %s", chartName))
		if repoURL == "" || chartName == "" {
			return nil, fmt.Errorf("missing helm metadata, metadata.RepoURL: %s, metadata.ChartName: %s", repoURL, chartName)
		}
		cfg := &repo.Entry{
			URL:  repoURL,
			Name: chartName,
		}
		cr, err := repo.NewChartRepository(cfg, getter.All(&cli.EnvSettings{}))
		if err != nil {
			return nil, err
		}
		repoBackend = &HelmRepoBackend{ChartRepo: cr}
	case S3HelmRepo:
		log.Debug(fmt.Sprintf("Configuring S3HelmRepoBackend for %s and repo %s", chartName, repoURL))

		bucketUrl, err := url.Parse(repoURL)
		if err != nil {
			return nil, err
		}
		repoBackend = &S3HelmRepoBackend{
			s3client:  s3Api,
			bucketUrl: bucketUrl,
		}
	case LocalRepo, GitRepo, GitRepoLocal:
		log.Debug(fmt.Sprintf("Chart url pointing to a repo of type %s. No chart helm repository backend to create for top-level chart", repoType))
		return nil, nil
	}
	return repoBackend, nil
}
