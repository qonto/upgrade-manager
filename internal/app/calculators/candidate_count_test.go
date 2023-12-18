package calculators

import (
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/stretchr/testify/assert"
	"log/slog"
)

func TestCandidateCountCalculateObsolescence(t *testing.T) {
	testCases := []struct {
		description   string
		soft          *software.Software
		expectedScore int
	}{
		{
			description: "many versions",
			soft: &software.Software{
				VersionCandidates: []software.Version{
					{Version: "1.0.0"},
					{Version: "1.1.0"},
					{Version: "1.2.0"},
				},
			},
			expectedScore: 3 * DefaultPerCandidateScore,
		},
		{
			description:   "no versions",
			soft:          &software.Software{},
			expectedScore: 0 * DefaultPerCandidateScore,
		},
	}
	calc := New(slog.Default(), software.CandidateCountCalculator, true)
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := calc.CalculateObsolescenceScore(tc.soft)
			assert.NoError(t, err)
			assert.Equal(t, tc.soft.CalculatedScore, tc.expectedScore)
		})
	}
}
