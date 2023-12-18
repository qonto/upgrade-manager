package calculators

import (
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
)

const DefaultPerCandidateScore = 30

type CandidateCountCalculator struct {
	checkDependencies bool
	perCandidateScore int
}

func (c *CandidateCountCalculator) CalculateObsolescenceScore(s *soft.Software) error {
	softwaresToCalculate := GetSoftwaresToCalculate(s, c.checkDependencies)
	for _, software := range softwaresToCalculate {
		software.CalculatedScore = len(s.VersionCandidates) * c.perCandidateScore
	}
	return nil
}
