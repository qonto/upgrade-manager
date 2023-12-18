package calculators

import (
	"sort"

	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
)

type ReleaseDateCalculator struct {
	checkDependencies bool
}

func (c *ReleaseDateCalculator) CalculateObsolescenceScore(s *soft.Software) error {
	softwaresToCalculate := GetSoftwaresToCalculate(s, c.checkDependencies)
	totalDaysLate := 0
	for _, software := range softwaresToCalculate {
		sort.Slice(software.VersionCandidates, func(i, j int) bool {
			v1 := software.VersionCandidates[i]
			v2 := software.VersionCandidates[j]
			return v2.ReleaseDate.Before(v1.ReleaseDate)
		})
		softwareDaysLate := 0
		currentDate := software.Version.ReleaseDate
		for _, candidate := range software.VersionCandidates {
			days := int(candidate.ReleaseDate.Sub(currentDate).Hours() / 24)
			if days > softwareDaysLate {
				softwareDaysLate = days
			}
		}
		totalDaysLate += softwareDaysLate
	}
	s.CalculatedScore = totalDaysLate * 5
	return nil
}
