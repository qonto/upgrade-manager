package argohelm

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/app/sources/helm/versions"
	"github.com/qonto/upgrade-manager/internal/app/sources/utils/gitutils"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
	"helm.sh/helm/v3/pkg/chart"
)

const ArgoHelm soft.SoftwareType = "argoHelm"

type Source struct {
	k8sClient          kubernetes.KubernetesClient
	log                *slog.Logger
	gitRepoConnections []*gitutils.RepoConnection
	cfg                Config
	s3Api              aws.S3Api
	versionFilter      filters.Filter
}

// Returns a new argohelm software source
func NewSource(cfg Config, log *slog.Logger, k8sClient kubernetes.KubernetesClient, loadSecretFromNamespace bool, s3Api aws.S3Api) (*Source, error) {
	if cfg.Filters.SemverVersions == nil {
		cfg.Filters.SemverVersions = &filters.SemverVersionsConfig{}
	}
	chartFilter := filters.Build(cfg.Filters.Config)
	s := &Source{
		log:           log,
		cfg:           cfg,
		k8sClient:     k8sClient,
		s3Api:         s3Api,
		versionFilter: chartFilter,
	}

	if loadSecretFromNamespace {
		if cfg.GitCredentialsSecretsPattern == "" {
			cfg.GitCredentialsSecretsPattern = ".*-repo-.*"
		}
		r := regexp.MustCompile(cfg.GitCredentialsSecretsPattern)
		connections, err := getGitRepoConnections(cfg.GitSecretsNamespace, r, k8sClient, log)
		if err != nil {
			s.log.Error(fmt.Sprintf("Failed to get Git Credential Secrets from namespace %s", cfg.GitSecretsNamespace))
			return nil, err
		}
		s.gitRepoConnections = connections
		log.Info(fmt.Sprintf("Found %d connections", len(connections)))
	}
	return s, nil
}

// TODO: fn on pointers
func (s *Source) Name() string {
	return "argohelm"
}

// Detects and provides a list of argocd helm applications as softwares
func (s *Source) Load() ([]*soft.Software, error) {
	var softwares []*soft.Software
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	filter, err := kubernetes.NewDestinationNamespaceFilter(s.cfg.Filters.FiltersOptions)
	if err != nil {
		return nil, err
	}
	apps, err := s.k8sClient.ListArgoApplications(ctx, s.cfg.ArgoCDNamespace, filter)
	if err != nil {
		return nil, err
	}
	s.log.Info(fmt.Sprintf("Found %d applications in %s namespace", len(apps), s.cfg.ArgoCDNamespace))
	for _, app := range apps {
		topLevelSoftware, chart, err := s.argoAppToSoftware(app)
		if err != nil {
			s.log.Warn(fmt.Sprintf("Could not convert argo app %s to software: %s", app.Name, err))
			continue
		}
		err = versions.PopulateTopLevelSoftware(s.s3Api, s.log, topLevelSoftware, app.RepoURL, app.Chart, s.versionFilter)
		if err != nil {
			s.log.Warn(fmt.Sprintf("Could not populate top level software %s: %s", app.Name, err))
		}
		if chart != nil {
			err = versions.PopulateSoftwareDependencies(s.s3Api, s.log, topLevelSoftware, chart, ArgoHelm, s.versionFilter)
			if err != nil {
				s.log.Error(fmt.Sprintf("Could not load %s chart as software, error: %s", chart.Name(), err))
				continue
			}
		}
		softwares = append(softwares, topLevelSoftware)
	}
	return softwares, nil
}

// Convert Argo app to Software
func (s *Source) argoAppToSoftware(app *kubernetes.ArgoCDApplication) (*soft.Software, *chart.Chart, error) {
	s.log.Debug(fmt.Sprintf("Converting argo app %s to software", app.Name))
	switch app.RepoBackendType { //nolint
	case versions.HelmRepo:
		// TODO: consider if we check dependencies or if we consider only the top-level chart in this case
		software := &soft.Software{
			Name:    app.Name,
			Version: soft.Version{Version: app.CurrentVersion},
			Type:    ArgoHelm,
		}
		s.log.Info(fmt.Sprintf("Adding software  %s ", software.Name))
		return software, nil, nil

	case versions.GitRepo:
		conn, err := s.matchGitRepoConnection(app.RepoURL)
		if err != nil {
			return nil, nil, err
		}
		dirName := os.TempDir() + "/tmp_" + app.Name + app.GitRevision + strconv.Itoa(int(time.Now().Unix()))
		chart, err := getChart(dirName, conn, app.RepoURL, app.ChartFilePath, app.GitRevision)
		if err != nil {
			return nil, nil, err
		}
		software := &soft.Software{
			Name: chart.Name(),
			Version: soft.Version{
				Version: versions.CleanVersion(chart.Metadata.Version),
			},
			Type: ArgoHelm,
		}
		return software, &chart, nil
	default:
		return nil, nil, fmt.Errorf("unexpected repo type %s", app.RepoBackendType)
	}
}

// Retrieve Chart.yaml from the chart remote RepoUrl (git)
func getChart(dirName string, conn gitutils.RepoConnectionProvider, remoteRepoUrl string, chartfileDir string, revision string) (chart.Chart, error) {
	r, err := conn.Clone(dirName, remoteRepoUrl, revision)
	if err != nil {
		return chart.Chart{}, err
	}
	w, err := r.Worktree()
	if err != nil {
		return chart.Chart{}, err
	}
	err = w.Checkout(&git.CheckoutOptions{Branch: plumbing.ReferenceName(revision), Create: true})
	if err != nil {
		return chart.Chart{}, err
	}
	var c chart.Chart
	if chartfileDir != "" {
		c, err = versions.LoadChartFile(fmt.Sprintf("%s/%s/%s", dirName, chartfileDir, "Chart.yaml"))
		if err != nil {
			return chart.Chart{}, err
		}
	} else {
		c, err = versions.LoadChartFile(fmt.Sprintf("%s/%s", dirName, "Chart.yaml"))
		if err != nil {
			return chart.Chart{}, err
		}
	}
	err = os.RemoveAll(dirName)
	if err != nil {
		return chart.Chart{}, err
	}
	return c, nil
}

// Retrieve the appropriate gitRepoConnection for the url
// based on repo url regex matching
func (s *Source) matchGitRepoConnection(url string) (*gitutils.RepoConnection, error) {
	// if url matches exactly an existing repo connection, use it

	for _, conn := range s.gitRepoConnections {
		if conn.Url == url {
			s.log.Debug(fmt.Sprintf("found matching repo credential for  %s", conn.Url))
			return conn, nil
		}
	}

	var foundCredentialTemplates []*gitutils.RepoConnection

	// else if url matches pattern of existing repo connection, use it
	for _, conn := range s.gitRepoConnections {
		r, err := regexp.Compile(conn.Url + ".*")
		if err != nil {
			return nil, err
		}
		if r.MatchString(url) {
			s.log.Debug(fmt.Sprintf("Found matching repo credential template  %s for  %s", conn.Url, url))
			foundCredentialTemplates = append(foundCredentialTemplates, conn)
		}
	}

	if len(foundCredentialTemplates) > 0 {
		if len(foundCredentialTemplates) > 1 {
			s.log.Warn(fmt.Sprintf("Found %d (more than 1) repo credential templates for %s. Using first one found...", len(foundCredentialTemplates), url))
		}
		return foundCredentialTemplates[0], nil
	}

	// else use a default unauthenticated repo backend
	s.log.Debug(fmt.Sprintf("using unauthenticated backend for %s", url))

	return &gitutils.RepoConnection{
		Auth: &http.BasicAuth{
			Username: "",
			Password: "",
		},
		Url:     url,
		Private: false,
	}, nil
}
