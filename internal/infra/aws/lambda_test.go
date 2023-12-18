package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
)

func TestGetDefaultDeprecatedRuntimesList(t *testing.T) {
	rt := GetDefaultDeprecatedRuntimesList()
	assert.NotEmpty(t, rt)
}

func TestIsLambdaRuntime(t *testing.T) {
	rts := ListLambdaRuntimes()
	expectedCount := len(rts)
	actualCount := 0

	notLambdaRuntimes := []types.Runtime{"perl", "cobol"}
	rts = append(rts, notLambdaRuntimes...)
	for _, rt := range rts {
		if IsLambdaRuntime(rt) {
			actualCount++
		}
	}
	assert.Equal(t, expectedCount, actualCount)
}

func TestFamilyVersionProvider(t *testing.T) {
	defaultValidateResult := func(t *testing.T, rts []types.Runtime) {
		t.Helper()
		assert.NotEmpty(t, rts)
	}
	testCases := []struct {
		Description    string
		Runtime        types.Runtime
		ValidateResult func(t *testing.T, rts []types.Runtime)
	}{
		{
			Description: "python",
			Runtime:     "python3.9",
		},
		{
			Description: "nodejs",
			Runtime:     "nodejs18.x",
		},
		{
			Description: "go",
			Runtime:     "go1.x",
		},
		{
			Description: "dotnetcore",
			Runtime:     "dotnetcore1.0",
		},
		{
			Description: "java",
			Runtime:     "java8",
		},
		{
			Description: "unknown runtime",
			Runtime:     "unknown",
			ValidateResult: func(t *testing.T, rts []types.Runtime) {
				t.Helper()
				assert.Empty(t, rts)
			},
		},
	}
	provideFamilyVersions := NewLambdaRuntimeFamilyVersionProvider()
	for i := range testCases {
		t.Run(testCases[i].Description, func(t *testing.T) {
			result := provideFamilyVersions(testCases[i].Runtime)
			if testCases[i].ValidateResult == nil {
				testCases[i].ValidateResult = defaultValidateResult
			}
			testCases[i].ValidateResult(t, result)
		})
	}
}
