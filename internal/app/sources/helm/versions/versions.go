package versions

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"log/slog"
)

func PopulateTopLevelSoftware(s3Api aws.S3Api, log *slog.Logger, topLevelSoftware *soft.Software, repoURL string, chartName string, filter filters.Filter) error {
	log.Debug(fmt.Sprintf("Populate top level software for chart %s repo backend %s", chartName, repoURL))
	repoBackend, err := buildRepoBackend(repoURL, chartName, log, s3Api)
	if err != nil {
		return err
	}
	if repoBackend == nil {
		return nil
	}
	err = computeSoftwareVersions(repoBackend, chartName, topLevelSoftware, filter, log)
	if err != nil {
		return err
	}
	return nil
}

func PopulateSoftwareDependencies(s3Api aws.S3Api, log *slog.Logger, topLevelSoftware *soft.Software, chart *chart.Chart, st soft.SoftwareType, filter filters.Filter) error {
	softwareDependencies := []*soft.Software{}
	for _, dependency := range chart.Metadata.Dependencies {
		var depName string
		if dependency.Alias != "" {
			depName = fmt.Sprintf("%s (%s)", dependency.Alias, dependency.Name)
		} else {
			depName = dependency.Name
		}
		softwareDependency := &soft.Software{
			Name: depName,
			Version: soft.Version{
				Version: CleanVersion(dependency.Version),
			},
			Type: st,
		}
		log.Debug(fmt.Sprintf("Populate dependency software for dep %s repo backend %s", depName, dependency.Repository))

		depRepoBackend, err := buildRepoBackend(dependency.Repository, dependency.Name, log, s3Api)
		if err != nil {
			return err
		}
		// we cannot do anything with this dependency TODO log ?
		if depRepoBackend == nil {
			continue
		}
		err = computeSoftwareVersions(depRepoBackend, dependency.Name, softwareDependency, filter, log)
		if err != nil {
			return err
		}
		softwareDependencies = append(softwareDependencies, softwareDependency)
	}

	topLevelSoftware.Dependencies = softwareDependencies
	return nil
}

func computeSoftwareVersions(repoBackend RepoBackend, chartName string, s *soft.Software, filter filters.Filter, log *slog.Logger) error {
	index, err := repoBackend.getIndexFile()
	if err != nil {
		return err
	}
	filteredVersions := []soft.Version{}

	allVersions, ok := index.Entries[chartName]
	if !ok {
		return fmt.Errorf("Fail to get chart versions for chart %s (repo %s)", chartName, repoBackend)
	}
	log.Debug(fmt.Sprintf("Found a total of %d versions for chart %s", allVersions.Len(), chartName))

	trustedCreationDate := trustIndexFileCreatedField(allVersions)
	log.Debug("Trusting index.yml file Created field")

	for i := range allVersions {
		chartVersion := allVersions[i]
		candidateVersion := software.Version{
			Name:    chartVersion.Name,
			Version: chartVersion.Version,
		}

		if trustedCreationDate {
			candidateVersion.ReleaseDate = chartVersion.Created
		} else {
			// prefilter charts
			// can fail for dates versions because not released yet
			keep := filter(s.Version, candidateVersion)
			if !keep {
				continue
			}
			releaseDate, err := getReleaseDateFromArchive(chartVersion, repoBackend)
			if err != nil {
				log.Warn(fmt.Sprintf("Fail to get release date from archive for chart %s version %s: %s", chartVersion.Name, chartVersion.Version, err.Error()))
				continue
			}
			candidateVersion.ReleaseDate = releaseDate
		}
		keep := filter(s.Version, candidateVersion)
		if keep {
			filteredVersions = append(filteredVersions, candidateVersion)
		}
	}

	log.Debug(fmt.Sprintf("Found a total of %d newer versions for chart %s", len(filteredVersions), chartName))
	s.VersionCandidates = filteredVersions
	return nil
}

// Defines if we should trust the index.yaml file Created field (known behavior: https://github.com/Talend/helm-charts-public/issues/5)
func trustIndexFileCreatedField(versions repo.ChartVersions) bool {
	for _, testVersion := range versions {
		yt, mt, dt := testVersion.Created.UTC().Date()
		sameDateRelease := 0
		// test if properly set created timestamp or if we need to check directly the tgz date
		// check the first 2 most recent versions vs all other versions
		for _, v := range versions {
			y, m, d := v.Created.UTC().Date()
			if y == yt && m == mt && d == dt {
				sameDateRelease++
			}
		}
		// if more than a third of the versions matched the testVersion
		if sameDateRelease > len(versions)/3 {
			return false
		}
	}
	return true
}

func CleanVersion(s string) string {
	cleanups := map[string]string{"^": "", "*": "0", "~": "", ">": "", "<": "", "=": "", "!": ""}
	for old, new := range cleanups {
		s = strings.ReplaceAll(s, old, new)
	}
	return s
}

// Retrieves the release date of the chart based on the last modification date of the actual chart .tgz file
func getReleaseDateFromArchive(v *repo.ChartVersion, backend RepoBackend) (time.Time, error) {
	url, err := url.Parse(v.URLs[0])
	if err != nil {
		return time.Now(), err
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	rawTarGZ, err := backend.getFile(ctx, url)
	if err != nil {
		return time.Now(), err
	}
	uncompressedStream, err := gzip.NewReader(rawTarGZ)
	if err != nil {
		return time.Now(), err
	}
	tarReader := tar.NewReader(uncompressedStream)
	entry, err := tarReader.Next()
	if err != nil {
		return time.Now(), err
	}
	return entry.FileInfo().ModTime(), nil
}
