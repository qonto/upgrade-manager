package calculators

import (
	"fmt"
	"testing"

	s "github.com/qonto/upgrade-manager/internal/app/core/software"
	"log/slog"
)

func TestMetaCalculateObsolescenceScore(t *testing.T) {
	testCases := []struct {
		software      *s.Software
		expectedScore int
	}{
		{
			software: &s.Software{
				Name:       "EKS-cluster-metacalculator",
				Calculator: s.MetaCalculator,
				Dependencies: []*s.Software{
					{
						Name:              "k8s",
						VersionCandidates: []s.Version{{Version: "1.24"}},
						Version:           s.Version{Version: "1.23"},
						Calculator:        s.AugmentedSemverCalculator,
					},
					{
						Name:              "ebs-addon",
						VersionCandidates: []s.Version{{Version: "1.3.4"}},
						Version:           s.Version{Version: "1.0.0"},
						Calculator:        s.SemverCalculator,
					},
				},
			},
			expectedScore: defaultMajorVersionScore + defaultMinorVersionScore*3,
		},
	}
	for _, tc := range testCases {
		fmt.Println(tc.software.Name)
		calculator := New(slog.Default(), tc.software.Calculator, true)
		err := calculator.CalculateObsolescenceScore(tc.software)
		if err != nil {
			t.Fatal(err)
		}
		if tc.software.CalculatedScore != tc.expectedScore {
			t.Fatalf("Expected score of %d, got: %d", tc.expectedScore, tc.software.CalculatedScore)
		}
	}
}
