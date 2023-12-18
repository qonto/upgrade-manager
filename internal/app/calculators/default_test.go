package calculators

import (
	"fmt"
	"testing"

	s "github.com/qonto/upgrade-manager/internal/app/core/software"
	"go.uber.org/zap"
)

func TestCalculateObsolescenceScore(t *testing.T) {
	testCases := []struct {
		software      *s.Software
		expectedScore int
	}{
		{
			software: &s.Software{
				Name:              "no-deps-and-up-to-date",
				Version:           s.Version{Version: "11.1.1"},
				VersionCandidates: []s.Version{},
				Calculator:        s.SemverCalculator,
			},
			expectedScore: 0,
		},
		{
			software: &s.Software{
				Name:              "deps and up-to-date",
				Version:           s.Version{Version: "11.1.1"},
				VersionCandidates: []s.Version{},
				Calculator:        s.SemverCalculator,
				Dependencies: []*s.Software{
					{
						Name:              "dep1",
						VersionCandidates: []s.Version{{Version: "17.0.1"}},
						Version:           s.Version{Version: "11.1.1"},
					},
					{
						Name:              "dep2",
						VersionCandidates: []s.Version{{Version: "3.0.1"}},
						Version:           s.Version{Version: "3.0.0"},
					},
				},
			},
			expectedScore: 6*defaultMajorVersionScore + defaultPatchVersionScore,
		},
	}
	for _, tc := range testCases {
		fmt.Println(tc.software.Name)
		calculator := New(zap.NewExample(), tc.software.Calculator, true)
		err := calculator.CalculateObsolescenceScore(tc.software)
		if err != nil {
			t.Fatal(err)
		}
		if tc.software.CalculatedScore != tc.expectedScore {
			t.Fatalf("Expected score of %d, got: %d", tc.expectedScore, tc.software.CalculatedScore)
		}
	}
}
