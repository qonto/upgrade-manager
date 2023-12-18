package msk

import "github.com/qonto/upgrade-manager/internal/app/filters"

type Config struct {
	Enabled        bool           `yaml:"enabled"`
	RequestTimeout string         `yaml:"request-timeout"`
	Filters        filters.Config `yaml:"filters"`
}
