package msk

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
)

type Source struct {
	api    aws.MSKApi
	log    *slog.Logger
	cfg    *Config
	filter filters.Filter
}

const (
	MskCluster     software.SoftwareType = "msk cluster"
	DefaultTimeout time.Duration         = time.Second * 15

	// From version 3.7.0, AWS start using 3.7.x version with automated patch update manage by themselves.
	// See https://docs.aws.amazon.com/msk/latest/developerguide/supported-kafka-versions.html#3.7.kraft
	//
	// When a candidate with x appears, we replace them with a 0 version to be still able to process them.
	SemVerWithX = `(\d+|x)\.(\d+|x)\.(\d+|x)`
)

func (s *Source) Name() string {
	return "MSK"
}

func NewSource(api aws.MSKApi, log *slog.Logger, cfg *Config) (*Source, error) {
	cfg.Filters = filters.Config{
		SemverVersions: &filters.SemverVersionsConfig{
			RemovePreRelease: true,
		},
	}
	filter := filters.Build(cfg.Filters)
	return &Source{
		log:    log,
		api:    api,
		cfg:    cfg,
		filter: filter,
	}, nil
}

func (s *Source) Load() ([]*software.Software, error) {
	softwares := []*software.Software{}
	res, err := s.api.ListClustersV2(context.TODO(), &kafka.ListClustersV2Input{})
	if err != nil {
		return nil, err
	}

	for _, cluster := range res.ClusterInfoList {
		res, err := s.api.GetCompatibleKafkaVersions(context.TODO(), &kafka.GetCompatibleKafkaVersionsInput{
			ClusterArn: cluster.ClusterArn,
		})
		if err != nil {
			return nil, err
		}

		versionCandidates := []software.Version{}
		for _, v := range res.CompatibleKafkaVersions[0].TargetVersions {
			versionCandidate := cleanMSKVersionSpecials(v)
			versionCandidates = append(versionCandidates, software.Version{Version: versionCandidate})
		}
		s := &software.Software{
			Calculator:        software.SemverCalculator,
			Name:              *cluster.ClusterName,
			Type:              MskCluster,
			Version:           software.Version{Version: *cluster.Provisioned.CurrentBrokerSoftwareInfo.KafkaVersion},
			VersionCandidates: versionCandidates,
		}
		softwares = append(softwares, s)
	}

	return softwares, nil
}

func cleanMSKVersionSpecials(version string) string {
	semverXRegexp := regexp.MustCompile(SemVerWithX)

	// Clean tiered version
	versionCandidate := strings.ReplaceAll(version, ".tiered", "")

	// Clean .x version to be valid semver
	return semverXRegexp.ReplaceAllStringFunc(versionCandidate, func(m string) string {
		return regexp.MustCompile(`x`).ReplaceAllString(m, "0")
	})
}
