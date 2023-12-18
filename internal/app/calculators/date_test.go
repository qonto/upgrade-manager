package calculators

import (
	"testing"
	"time"

	s "github.com/qonto/upgrade-manager/internal/app/core/software"
	"log/slog"
)

func TestDateCalculateObsolescenceScore(t *testing.T) {
	testCases := []struct {
		software      *s.Software
		expectedScore int
	}{
		{
			software: &s.Software{
				Version:           s.Version{ReleaseDate: time.Now()},
				VersionCandidates: []s.Version{},
				Dependencies: []*s.Software{
					{
						VersionCandidates: []s.Version{{ReleaseDate: time.Now()}},
						Version:           s.Version{ReleaseDate: time.Now().Add(-49 * time.Hour)},
					},
					{
						VersionCandidates: []s.Version{{ReleaseDate: time.Now()}},
						Version:           s.Version{ReleaseDate: time.Now().Add(-73 * time.Hour)},
					},
				},
			},
			expectedScore: 25,
		},
		{
			software: &s.Software{
				Version:           s.Version{ReleaseDate: time.Now()},
				VersionCandidates: []s.Version{},
				Dependencies: []*s.Software{
					{
						VersionCandidates: []s.Version{
							{ReleaseDate: time.Now().Add(-48 * time.Hour)},
							{ReleaseDate: time.Now()},
						},
						Version: s.Version{ReleaseDate: time.Now().Add(-49 * time.Hour)},
					},
					{
						VersionCandidates: []s.Version{{ReleaseDate: time.Now()}},
						Version:           s.Version{ReleaseDate: time.Now().Add(-73 * time.Hour)},
					},
				},
			},
			expectedScore: 25,
		},
		{
			software: &s.Software{
				Version: s.Version{ReleaseDate: time.Now().Add(-73 * time.Hour)},
				VersionCandidates: []s.Version{
					{ReleaseDate: time.Now().Add(-23 * time.Hour)},
				},
			},
			expectedScore: 10,
		},
	}
	calculator := New(slog.Default(), s.ReleaseDateCalculator, true)
	for _, tc := range testCases {
		err := calculator.CalculateObsolescenceScore(tc.software)
		if err != nil {
			t.Fatal(err)
		}
		if tc.software.CalculatedScore != tc.expectedScore {
			t.Fatalf("Expected score of %d, got: %d", tc.expectedScore, tc.software.CalculatedScore)
		}
	}
}
