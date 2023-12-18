package lambda

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type Config struct {
	Enabled                 bool            `yaml:"enabled"`
	RequestTimeout          string          `yaml:"request-timeout"`
	DeprecatedRuntimes      []types.Runtime `yaml:"deprecated-runtimes"`
	DeprecatedRuntimesScore int             `yaml:"deprecated-runtimes-score"`
}
