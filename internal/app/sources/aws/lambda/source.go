package lambda

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
)

type Source struct {
	log *slog.Logger
	api aws.LambdaApi
	cfg *Config
}

const (
	LambdaFunction                 software.SoftwareType = "lambda"
	DefaultTimeout                 time.Duration         = time.Second * 15
	DefaultDeprecatedRuntimesScore int                   = 100
)

func (s *Source) Name() string {
	return "lambda"
}

func NewSource(api aws.LambdaApi, log *slog.Logger, cfg *Config) (*Source, error) {
	return &Source{
		api: api,
		log: log,
		cfg: cfg,
	}, nil
}

func (s *Source) Load() ([]*software.Software, error) {
	var softwares []*software.Software
	timeout, err := time.ParseDuration(s.cfg.RequestTimeout)
	if err != nil || s.cfg.RequestTimeout == "" {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	rtVersionProvider := aws.NewLambdaRuntimeFamilyVersionProvider()
	var deprecatedRuntimes []types.Runtime
	if s.cfg.DeprecatedRuntimes != nil {
		deprecatedRuntimes = s.cfg.DeprecatedRuntimes
	} else {
		deprecatedRuntimes = aws.GetDefaultDeprecatedRuntimesList()
	}

	res, err := s.api.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, err
	}
	for _, f := range res.Functions {
		soft := &software.Software{
			Name:    *f.FunctionName,
			Version: *software.ToVersion(string(f.Runtime)),
			Type:    LambdaFunction,
		}

		if !aws.IsLambdaRuntime(f.Runtime) {
			continue
		}

		// set arbitrary score if the runtime is deprecated
		for _, deprecrated := range deprecatedRuntimes {
			if f.Runtime == deprecrated {
				soft.Calculator = software.SkipCalculator
				if s.cfg.DeprecatedRuntimesScore == 0 {
					soft.CalculatedScore = DefaultDeprecatedRuntimesScore
				} else {
					soft.CalculatedScore = s.cfg.DeprecatedRuntimesScore
				}
			}
		}

		if soft.CalculatedScore == 0 {
			soft.Calculator = software.CandidateCountCalculator
		}

		// load version candidates
		familyRuntimes := rtVersionProvider(f.Runtime)
		var found bool
		for i := range familyRuntimes {
			found = false
			if familyRuntimes[i] == f.Runtime {
				found = true
				candidates := familyRuntimes[i+1:]
				for _, rt := range candidates {
					soft.VersionCandidates = append(soft.VersionCandidates, *software.ToVersion(string(rt)))
				}
				break
			}
		}
		// If we did not find the function's runtime version in the list of supported versions,
		// then all the runtime versions in the runtime family are candidates
		if !found {
			for _, rt := range familyRuntimes {
				soft.VersionCandidates = append(soft.VersionCandidates, *software.ToVersion(string(rt)))
			}
		}

		// Reverse the slice to provide the most recent candidate as a first element (aws lambda version does not follow semver or another stable versioning logic)
		// ex:
		// python family: "python2.7", "python3.6", "python3.7" etc...
		// nodejs family: "nodejs12.x", "nodejs", "nodejs4.3-edge", etc...
		// java family: "java8", "java11", "java.al2", etc...

		for i, j := 0, len(soft.VersionCandidates)-1; i < j; i, j = i+1, j-1 {
			soft.VersionCandidates[i], soft.VersionCandidates[j] = soft.VersionCandidates[j], soft.VersionCandidates[i]
		}

		s.log.Info("Tracking software", slog.String("software", soft.Name), slog.String("software_type", string(soft.Type)))
		softwares = append(softwares, soft)
	}
	return softwares, nil
}
