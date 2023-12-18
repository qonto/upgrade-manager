package eks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/app/semver"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
)

type Source struct {
	log    *slog.Logger
	api    aws.EKSApi
	cfg    *Config
	filter filters.Filter
}

const (
	EksCluster     software.SoftwareType = "eks-cluster"
	K8s            software.SoftwareType = "k8s-engine"
	EksAddon       software.SoftwareType = "eks-addon"
	DefaultTimeout time.Duration         = time.Second * 15
)

func (s *Source) Name() string {
	return "EKS"
}

func NewSource(api aws.EKSApi, log *slog.Logger, cfg *Config) (*Source, error) {
	// Current implementation of filters requires this map to be non-nil to filter old versions
	// so we set RemovePreRelease to true to filter out old versions anyway.
	// NOTE: this is slightly confusing and should probably be refactored later on
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
	timeout, err := time.ParseDuration(s.cfg.RequestTimeout)
	if err != nil || s.cfg.RequestTimeout == "" {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	softwares := []*software.Software{}
	res, err := s.api.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		s.log.Error(fmt.Sprintf("%s", err))
		return nil, err
	}

	for i, clusterName := range res.Clusters {
		topLevelSoft := &software.Software{
			Type:       EksCluster,
			Name:       clusterName,
			Calculator: software.MetaCalculator,
		}

		// get cluster current version
		clusterInfo, err := s.api.DescribeCluster(ctx, &eks.DescribeClusterInput{
			Name: &res.Clusters[i],
		})
		if err != nil {
			return nil, err
		}

		// get k8s latest released version
		k8sVersion, err := s.getK8sLatestVersion(ctx)
		if err != nil {
			return nil, err
		}
		s.log.Debug(fmt.Sprintf("Latest EKS version found for cluster '%s': %s ", clusterName, k8sVersion.Version))

		topLevelSoft.Dependencies = append(topLevelSoft.Dependencies, &software.Software{
			Type:              K8s,
			Name:              "k8s",
			VersionCandidates: []software.Version{*k8sVersion},
			Version:           software.Version{Version: *clusterInfo.Cluster.Version},
			Calculator:        software.AugmentedSemverCalculator,
		})

		// load cluster addon dependencies
		addons, err := s.api.ListAddons(ctx, &eks.ListAddonsInput{ClusterName: &res.Clusters[i]})
		if err != nil {
			return nil, err
		}
		for j, addon := range addons.Addons {
			// get currently deployed addon version
			currentAddonConfig, err := s.api.DescribeAddon(ctx, &eks.DescribeAddonInput{
				AddonName:   &addons.Addons[j],
				ClusterName: &res.Clusters[i],
			})
			if err != nil {
				return nil, err
			}

			currentAddonSemver, err := semver.ExtractFromString(*currentAddonConfig.Addon.AddonVersion)
			if err != nil {
				continue
			}
			addonDep := software.Software{
				Name:       addon,
				Calculator: software.SemverCalculator,
				Version:    software.Version{Version: currentAddonSemver},
			}

			// get all released version for the addon
			versions, err := s.getAddonVersions(ctx, addon, *clusterInfo.Cluster.Version)
			if err != nil {
				return nil, err
			}

			// filter candidates
			for _, v := range versions {
				if keep := s.filter(addonDep.Version, *v); keep {
					addonDep.VersionCandidates = append(addonDep.VersionCandidates, *v)
				}
			}
			topLevelSoft.Dependencies = append(topLevelSoft.Dependencies, &addonDep)
		}
		softwares = append(softwares, topLevelSoft)
		s.log.Info(fmt.Sprintf("Tracking software %s, of type %s", topLevelSoft.Name, topLevelSoft.Type))
	}
	return softwares, nil
}

// Returns all available versions given an EKS Addon name
func (s *Source) getAddonVersions(ctx context.Context, name string, clusterVersion string) (software.Versions, error) {
	versions := software.Versions{}
	res, err := s.api.DescribeAddonVersions(ctx, &eks.DescribeAddonVersionsInput{KubernetesVersion: &clusterVersion})
	if err != nil {
		return versions, err
	}
	for _, item := range res.Addons {
		if *item.AddonName == name {
			cleanAddonVersion := software.Versions{}
			for _, v := range item.AddonVersions {
				addonVersion, err := semver.ExtractFromString(*v.AddonVersion)
				if err != nil {
					return nil, err
				}
				cleanAddonVersion = append(cleanAddonVersion, &software.Version{
					Version: addonVersion,
				})
			}
			cleanAddonVersion.Deduplicate()
			versions = append(versions, cleanAddonVersion...)
		}
	}
	return versions, err
}

// Return latest k8s version offered by AWS
func (s *Source) getK8sLatestVersion(ctx context.Context) (*software.Version, error) {
	addonDesc, err := s.api.DescribeAddonVersions(ctx, &eks.DescribeAddonVersionsInput{})
	if err != nil {
		return nil, err
	}
	k8sVersion := &software.Version{
		Version: *addonDesc.Addons[0].AddonVersions[0].Compatibilities[0].ClusterVersion,
	}

	return k8sVersion, nil
}
