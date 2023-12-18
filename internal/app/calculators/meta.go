package calculators

import (
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"go.uber.org/zap"
)

var calculatorCache = make(map[software.CalculatorType]software.Calculator)

type MetaCalculator struct {
	log               *zap.Logger
	checkDependencies bool
}

// Entrypoint calculator which supports softwares with
// dependencies having different Calculator Types
func (c *MetaCalculator) CalculateObsolescenceScore(s *software.Software) error {
	softwaresToCompute := GetSoftwaresToCalculate(s, true)
	for _, soft := range softwaresToCompute {
		var sCalculator software.Calculator

		if existingCalculator, ok := calculatorCache[soft.Calculator]; ok {
			sCalculator = existingCalculator
		} else {
			calculatorCache[soft.Calculator] = New(c.log, soft.Calculator, false)
			sCalculator = calculatorCache[soft.Calculator]
		}

		if err := sCalculator.CalculateObsolescenceScore(soft); err != nil {
			return err
		}
	}
	if s.CalculatedScore == 0 {
		for _, dep := range s.Dependencies {
			s.CalculatedScore += dep.CalculatedScore
		}
	}
	return nil
}
