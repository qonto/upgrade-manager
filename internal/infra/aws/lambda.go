package aws

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type LambdaApi interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
}

// There is no API to retrieve a list of available runtimes.
//
// As a workaround, we source them from the lambda/types packages
// To fetch latest versions available
func ListLambdaRuntimes() []types.Runtime {
	return types.RuntimePython39.Values()
}

func ListSupportedRuntimes() []types.Runtime {
	supportedRuntimes := []types.Runtime{}
	rtList := ListLambdaRuntimes()
	drtList := GetDefaultDeprecatedRuntimesList()
	for _, rt := range rtList {
		supported := true
		for _, drt := range drtList {
			if rt == drt {
				supported = false
				break
			}
		}
		if supported {
			supportedRuntimes = append(supportedRuntimes, rt)
		}
	}
	return supportedRuntimes
}

func IsLambdaRuntime(runtime types.Runtime) bool {
	rtList := ListLambdaRuntimes()
	for _, rt := range rtList {
		if rt == runtime {
			return true
		}
	}
	return false
}

func GetDefaultDeprecatedRuntimesList() []types.Runtime {
	return []types.Runtime{
		"nodejs",
		"nodejs4.3",
		"nodejs4.3-edge",
		"nodejs6.10",
		"nodejs8.10",
		"nodejs10.x",
		"nodejs12.x",
		"java8",
		"java8.al2",
		"python2.7",
		"python3.6",
		"dotnetcore1.0",
		"dotnetcore2.0",
		"dotnet6",
		"ruby2.5",
	}
}

// Returns a function that, given a runtime, returns all the runtimes of the same family
func NewLambdaRuntimeFamilyVersionProvider() func(runtime types.Runtime) []types.Runtime {
	rtList := ListSupportedRuntimes()
	rtVersionByFamily := map[string][]types.Runtime{}
	matchers := map[string]*regexp.Regexp{
		"python":     regexp.MustCompile("python.+"),
		"nodejs":     regexp.MustCompile("nodejs.+"),
		"java":       regexp.MustCompile("java.+"),
		"dotnetcore": regexp.MustCompile("dotnetcore.+"),
		"dotnet":     regexp.MustCompile("dotnet.+"),
		"ruby":       regexp.MustCompile("ruby.+"),
		"go":         regexp.MustCompile("go.+"),
	}
	for _, rt := range rtList {
		for family, expr := range matchers {
			if expr.MatchString(string(rt)) {
				rtVersionByFamily[family] = append(rtVersionByFamily[family], rt)
			}
		}
	}
	return func(runtime types.Runtime) []types.Runtime {
		for family, expr := range matchers {
			if expr.MatchString(string(runtime)) {
				return rtVersionByFamily[family]
			}
		}
		return nil
	}
}
