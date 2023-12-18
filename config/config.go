package config

import (
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/qonto/upgrade-manager/internal/app/sources/argohelm"
	awsSource "github.com/qonto/upgrade-manager/internal/app/sources/aws"
	"github.com/qonto/upgrade-manager/internal/app/sources/deployments"
	"github.com/qonto/upgrade-manager/internal/app/sources/filesystemhelm"
	awsinfra "github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/qonto/upgrade-manager/internal/infra/http"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Global  GlobalConfig `yaml:"global" validate:"required"`
	Sources Sources      `yaml:"sources" validate:"required"`
	HTTP    http.Config  `yaml:"http" validate:"required"`
}
type GlobalConfig struct {
	Interval  string          `yaml:"interval" validate:"required"`
	AwsConfig awsinfra.Config `yaml:"aws" validate:"required"`
}

type Sources struct {
	Deployments []deployments.Config    `yaml:"deployments"`
	ArgocdHelm  []argohelm.Config       `yaml:"argocdHelm"`
	FsHelm      []filesystemhelm.Config `yaml:"filesystemHelm"`
	Aws         awsSource.Config        `yaml:"aws"`
}

// Load and unmarshal config file
func Load(configFilePath string) (Config, error) {
	var cfg Config
	f, err := os.ReadFile(configFilePath)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return cfg, err
	}
	_, err = time.ParseDuration(cfg.Global.Interval)
	if err != nil {
		return cfg, err
	}
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
