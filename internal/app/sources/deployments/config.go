package deployments

import (
	"github.com/qonto/upgrade-manager/internal/app/filters"
	"github.com/qonto/upgrade-manager/internal/infra/registry"
)

type Config struct {
	Name          string                     `yaml:"name"`
	Namespace     string                     `yaml:"namespace"`
	LabelSelector map[string]string          `yaml:"label-selector"`
	Registries    map[string]registry.Config `yaml:"registries"`
	Filters       filters.Config             `yaml:"filters"`
}
