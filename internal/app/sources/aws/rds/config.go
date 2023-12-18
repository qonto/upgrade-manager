package rds

import "github.com/qonto/upgrade-manager/internal/app/filters"

type Config struct {
	Enabled          bool           `yaml:"enabled"`
	AggregationLevel string         `yaml:"aggregation-level"` // cluster or instance
	RequestTimeout   string         `yaml:"request-timeout"`
	Filters          filters.Config `yaml:"filters"`
}
