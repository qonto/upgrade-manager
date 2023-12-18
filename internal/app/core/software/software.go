package software

import (
	"fmt"
)

type SoftwaresToUpdate struct {
	Softwares []Software
}

type CalculatorType string

const (
	// Augmented Semver: for softwares where minor versions are considered major versions (ex: kubernetes 1.21 -> 1.22)
	AugmentedSemverCalculator CalculatorType = "augmented-semver"
	CandidateCountCalculator  CalculatorType = "candidate-count"
	MetaCalculator            CalculatorType = "meta"
	ReleaseDateCalculator     CalculatorType = "release-date"
	SemverCalculator          CalculatorType = "semver"
	SkipCalculator            CalculatorType = "skip"
)

type SoftwareType string

type Software struct {
	Calculator        CalculatorType
	CalculatedScore   int
	Dependencies      []*Software
	Name              string
	Type              SoftwareType
	Version           Version
	VersionCandidates []Version
}

func ToCalculator(name string) (CalculatorType, error) {
	switch name {
	case string(SkipCalculator):
		return SkipCalculator, nil
	case string(AugmentedSemverCalculator):
		return AugmentedSemverCalculator, nil
	case string(CandidateCountCalculator):
		return CandidateCountCalculator, nil
	case string(MetaCalculator):
		return MetaCalculator, nil
	case string(ReleaseDateCalculator):
		return ReleaseDateCalculator, nil
	case string(SemverCalculator):
		return SemverCalculator, nil
	default:
		return "", fmt.Errorf("Unknown calculator %s", name)
	}
}
