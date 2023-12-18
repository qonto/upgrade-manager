package filesystemhelm

import (
	"github.com/qonto/upgrade-manager/internal/app/filters"
)

type Config struct {
	Enabled bool           `yaml:"enabled"`
	Paths   []string       `yaml:"paths" validate:"dive,file"`
	Filters filters.Config `yaml:"filters"`
}
