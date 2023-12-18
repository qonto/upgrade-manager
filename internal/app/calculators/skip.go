package calculators

import (
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
)

// Accepts the arbitrarily set score by the source
type SkipCalculator struct{}

func (c *SkipCalculator) CalculateObsolescenceScore(s *soft.Software) error {
	return nil
}
