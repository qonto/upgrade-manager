package filesystemhelm

import (
	"fmt"

	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/app/sources/helm/versions"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"helm.sh/helm/v3/pkg/chart"
	"log/slog"
)

const FileSystemHelm soft.SoftwareType = "filesystemHelm"

type Source struct {
	Charts []chart.Chart
	log    *slog.Logger
	cfg    Config
	s3Api  aws.S3Api
	filter filters.Filter
}

func NewSource(cfg Config, log *slog.Logger, s3Api aws.S3Api) (*Source, error) {
	if cfg.Filters.SemverVersions == nil {
		cfg.Filters.SemverVersions = &filters.SemverVersionsConfig{}
	}
	chartFilter := filters.Build(cfg.Filters)

	s := Source{
		log:    log,
		cfg:    cfg,
		s3Api:  s3Api,
		filter: chartFilter,
	}
	// is the source initialized properly
	if cfg.Enabled {
		for _, chartFilePath := range cfg.Paths {
			// Move in the Load() function
			c, err := versions.LoadChartFile(chartFilePath)
			if err != nil {
				return &s, err
			}
			s.Charts = append(s.Charts, c)
		}
	}
	return &s, nil
}

func (s *Source) Load() ([]*soft.Software, error) {
	softwares := make([]*soft.Software, 0, len(s.Charts))
	// for chart.yaml we found
	for i := range s.Charts {
		chart := s.Charts[i]
		topLevelSoftware := &soft.Software{
			Name: chart.Name(),
			Version: soft.Version{
				Version: versions.CleanVersion(chart.Metadata.Version),
			},
			Type: FileSystemHelm,
		}

		err := versions.PopulateSoftwareDependencies(s.s3Api, s.log, topLevelSoftware, &chart, FileSystemHelm, s.filter)
		if err != nil {
			s.log.Error(fmt.Sprintf("Could not load %s chart as software, error: %s", chart.Name(), err))
			continue
		}
		softwares = append(softwares, topLevelSoftware)
	}
	return softwares, nil
}

func (s *Source) Name() string {
	return "fshelm"
}
