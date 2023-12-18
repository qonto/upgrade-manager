package lambda

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/sources/utils"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoad(t *testing.T) {
	defaultBuildMock := func(mockApi *aws.LambdaMock) {
		mockApi.On("ListFunctions", mock.Anything).Return(
			&lambda.ListFunctionsOutput{
				Functions: []types.FunctionConfiguration{
					{
						FunctionName: utils.Ptr("function1 - up to date"),
						Runtime:      types.RuntimeJava21,
					},
					{
						FunctionName: utils.Ptr("function2 - not deprecated but 2 versions late"),
						Runtime:      types.RuntimePython37,
					},
					{
						FunctionName: utils.Ptr("function3 - deprecated"),
						Runtime:      types.RuntimeNodejs12x,
					},
					{
						FunctionName: utils.Ptr("function4 - NotaRealRuntime"),
						Runtime:      "unknown runtime",
					},
				},
			}, nil,
		)
	}
	testCases := []struct {
		Description    string
		Config         *Config
		BuildMock      func(mockApi *aws.LambdaMock)
		ValidateResult func(t *testing.T, result []*software.Software, resultErr error)
	}{
		{
			Description: "default deprecrated runtimes list",
			Config: &Config{
				Enabled:                 true,
				DeprecatedRuntimesScore: DefaultDeprecatedRuntimesScore,
			},
			ValidateResult: func(t *testing.T, result []*software.Software, resultErr error) {
				t.Helper()
				assert.NoError(t, resultErr)
				assert.Len(t, result, 3)
				assert.Equal(t, software.CandidateCountCalculator, result[0].Calculator)
				assert.Empty(t, result[0].VersionCandidates)
				assert.Equal(t, software.CandidateCountCalculator, result[1].Calculator)
				assert.Len(t, result[1].VersionCandidates, 5)
				assert.Equal(t, software.SkipCalculator, result[2].Calculator)
				assert.Equal(t, DefaultDeprecatedRuntimesScore, result[2].CalculatedScore)
			},
		},
		{
			Description: "custom deprecated runtimes list",
			Config: &Config{
				Enabled:                 true,
				DeprecatedRuntimes:      append(aws.GetDefaultDeprecatedRuntimesList(), types.RuntimePython37),
				DeprecatedRuntimesScore: 150,
			},
			ValidateResult: func(t *testing.T, result []*software.Software, resultErr error) {
				t.Helper()
				assert.NoError(t, resultErr)
				assert.Len(t, result, 3)
				assert.Equal(t, software.CandidateCountCalculator, result[0].Calculator)
				assert.Empty(t, result[0].VersionCandidates)
				assert.Equal(t, software.SkipCalculator, result[1].Calculator)
				assert.Equal(t, 150, result[2].CalculatedScore)
				assert.Equal(t, software.SkipCalculator, result[2].Calculator)
				assert.Equal(t, 150, result[2].CalculatedScore)
			},
		},
		{
			Description: "failed call to ListFunction",
			Config: &Config{
				Enabled:            true,
				DeprecatedRuntimes: append(aws.GetDefaultDeprecatedRuntimesList(), types.RuntimePython38),
			},
			BuildMock: func(mockApi *aws.LambdaMock) {
				mockApi.On("ListFunctions", mock.Anything).Return(
					&lambda.ListFunctionsOutput{}, errors.New("failed to list functions"),
				)
			},
			ValidateResult: func(t *testing.T, result []*software.Software, resultErr error) {
				t.Helper()
				assert.Error(t, resultErr)
				assert.Empty(t, result)
			},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].Description, func(t *testing.T) {
			mockApi := new(aws.LambdaMock)
			if testCases[i].BuildMock == nil {
				testCases[i].BuildMock = defaultBuildMock
			}
			testCases[i].BuildMock(mockApi)
			source, err := NewSource(mockApi, slog.Default(), testCases[i].Config)

			assert.NoError(t, err)
			softwares, err := source.Load()
			testCases[i].ValidateResult(t, softwares, err)
			assert.NotEmpty(t, source.Name())
		})
	}
}
