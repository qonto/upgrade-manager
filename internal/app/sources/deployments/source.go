package deployments

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
	"github.com/qonto/upgrade-manager/internal/infra/registry"
)

const (
	Deployments soft.SoftwareType = "kubernetes-deployment"

	calculatorAnnotation = "upgrade-manager.qonto.com/calculator"
)

type Source struct {
	k8sClient             kubernetes.KubernetesClient
	defaultRegistryClient *registry.Client
	registryClients       map[string]*registry.Client
	log                   *slog.Logger
	cfg                   Config
	filter                filters.Filter
}

// TODO: fn on pointers
func (s *Source) Name() string {
	return "deployments"
}

func NewSource(log *slog.Logger, k8sClient kubernetes.KubernetesClient, cfg Config) (*Source, error) {
	filter := filters.Build(cfg.Filters)
	s := &Source{
		log:       log,
		cfg:       cfg,
		k8sClient: k8sClient,
		filter:    filter,
	}
	registryClients := make(map[string]*registry.Client)
	defaultRegistryClient, err := registry.New(&registry.Config{})
	if err != nil {
		return nil, err
	}
	s.defaultRegistryClient = defaultRegistryClient

	for registryName := range cfg.Registries {
		registryConfig := cfg.Registries[registryName]
		registryClient, err := registry.New(&registryConfig)
		if err != nil {
			return nil, err
		}
		registryClients[registryName] = registryClient
	}
	s.registryClients = registryClients
	return s, nil
}

type image struct {
	Repository string
	Version    string
	Registry   string
}

func buildImage(s string) (image, error) {
	dotSplitted := strings.Split(s, ":")
	if len(dotSplitted) < 2 {
		return image{}, fmt.Errorf("Invalid container image %s", s)
	}

	return image{
		Repository: dotSplitted[0],
		Version:    dotSplitted[1],
		Registry:   strings.Split(s, "/")[0],
	}, nil
}

func (s *Source) Load() ([]*soft.Software, error) {
	var softwares []*soft.Software
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	deployments, err := s.k8sClient.ListDeployments(ctx, kubernetes.ListRequest{Namespace: s.cfg.Namespace, LabelSelector: s.cfg.LabelSelector})
	if err != nil {
		return nil, err
	}
	s.log.Info(fmt.Sprintf("Found %d deployments", len(deployments.Items)))
	for _, deployment := range deployments.Items {
		name := deployment.Name
		containers := deployment.Spec.Template.Spec.Containers

		calculatorAnnotation, ok := deployment.Annotations[calculatorAnnotation]
		calculator := software.SemverCalculator
		if ok {
			calculator, err = soft.ToCalculator(calculatorAnnotation)
			if err != nil {
				s.log.Error(fmt.Sprintf("Invalid calculator for deployment %s: %s", name, err.Error()))
				continue
			}
		}
		dependencies := []*soft.Software{}
		for _, container := range containers {
			image, err := buildImage(container.Image)
			if err != nil {
				s.log.Error(err.Error())
				continue
			}
			registryClient, ok := s.registryClients[image.Registry]
			if !ok {
				s.log.Debug(fmt.Sprintf("Using default registry client for image %s", image))
				registryClient = s.defaultRegistryClient
			}
			tags, err := registryClient.Tags(image.Repository)
			if err != nil {
				s.log.Error(fmt.Sprintf("Fail to retrieve tags for repository %s: %s", image.Repository, err.Error()))
				continue
			}
			containerSoftware := &soft.Software{
				Calculator: calculator,
				Name:       fmt.Sprintf("%s-%s", name, container.Name),
				Version:    soft.Version{Version: image.Version},
				Type:       Deployments,
			}
			if registryClient.ReleaseDateRetrievalEnabled() {
				configFile, err := registryClient.ConfigFile(container.Image)
				if err != nil {
					s.log.Error(fmt.Sprintf("Fail to retrieve release date for image %s: %s", container.Image, err.Error()))
					continue
				}
				containerSoftware.Version.ReleaseDate = configFile.Created.Time
			}
			versionCandidates := []soft.Version{}
			for _, tag := range tags {
				if calculator == software.SemverCalculator {
					_, err := goversion.NewSemver(tag)
					if err != nil {
						s.log.Debug(fmt.Sprintf("Skipping non-semver version %s for image %s", tag, image.Repository))
						continue
					}
				}
				versionCandidate := soft.Version{Version: tag}
				// optimization to filter semver versions before retrieving creation date
				keep := s.filter(containerSoftware.Version, versionCandidate)
				if !keep {
					continue
				}
				if registryClient.ReleaseDateRetrievalEnabled() {
					depImage := fmt.Sprintf("%s:%s", image.Repository, tag)
					configFile, err := registryClient.ConfigFile(depImage)
					if err != nil {
						s.log.Error(fmt.Sprintf("Fail to retrieve release date for image %s: %s", depImage, err.Error()))
						continue
					}
					versionCandidate.ReleaseDate = configFile.Created.Time
				}
				keep = s.filter(containerSoftware.Version, versionCandidate)
				if keep {
					versionCandidates = append(versionCandidates, versionCandidate)
				}
			}
			s.log.Debug(fmt.Sprintf("Found %d version candidates for repository %s", len(versionCandidates), image.Repository))
			containerSoftware.VersionCandidates = versionCandidates
			dependencies = append(dependencies, containerSoftware)
		}
		softwares = append(softwares, &soft.Software{
			Calculator:   calculator,
			Name:         name,
			Type:         Deployments,
			Dependencies: dependencies,
		})
	}
	return softwares, nil
}
