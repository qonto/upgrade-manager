package argohelm

import (
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
)

type Config struct {
	Enabled                      bool          `yaml:"enabled"`
	Name                         string        `yaml:"name"`
	ClusterURL                   string        `yaml:"cluster-url"`
	ArgoCDNamespace              string        `yaml:"argocd-namespace" validate:"required"`
	GitSecretsNamespace          string        `yaml:"git-credentials-secrets-namespace" validate:"required"`
	GitCredentialsSecretsPattern string        `yaml:"git-credentials-secrets-pattern" validate:"required"`
	Filters                      FiltersConfig `yaml:"filters"`
}

type FiltersConfig struct {
	filters.Config            `yaml:",inline"`
	kubernetes.FiltersOptions `yaml:",inline"`
}
