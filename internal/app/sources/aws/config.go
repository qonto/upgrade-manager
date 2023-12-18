package awsSource

import (
	"github.com/qonto/upgrade-manager/internal/app/sources/aws/eks"
	"github.com/qonto/upgrade-manager/internal/app/sources/aws/elasticache"
	"github.com/qonto/upgrade-manager/internal/app/sources/aws/lambda"
	"github.com/qonto/upgrade-manager/internal/app/sources/aws/msk"
	"github.com/qonto/upgrade-manager/internal/app/sources/aws/rds"
)

type Config struct {
	Elasticache elasticache.Config `yaml:"elasticache"`
	Eks         eks.Config         `yaml:"eks"`
	Rds         rds.Config         `yaml:"rds"`
	Lambda      lambda.Config      `yaml:"lambda"`
	Msk         msk.Config         `yaml:"msk"`
}
