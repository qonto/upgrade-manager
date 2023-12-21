package calculators

import (
	"fmt"
	"log/slog"

	goversion "github.com/hashicorp/go-version"
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/semver"
)

const (
	defaultMajorVersionScore = 50
	defaultMinorVersionScore = 5
	defaultPatchVersionScore = 1
)

// Each calculator can come up with its own logic for computing
// the Obsolescence score

// The default calculator parses versions using Semantic Versioning.
// It takes the last version available and computes the obsolescence score:
//
// x major versions late = x * 50
//
// y minor versions late = y * 5
//
// z patch versions late = z * 1
type DefaultCalculator struct {
	log               *slog.Logger
	scoreTable        segmentScoreTable
	checkDependencies bool
}

// Mapping of score to allocate per version segment
type segmentScoreTable []int

// Provides a based on input type. Falls back on DefaultCalculator
//
// NOTE: Should DefaultCalculator be renamed to SemverCalculator
// and MetaCalculator be renamed to DefaultCalculator?
func New(logger *slog.Logger, t soft.CalculatorType, checkDependencies bool) soft.Calculator {
	switch t { //nolint
	case soft.ReleaseDateCalculator:
		return &ReleaseDateCalculator{
			checkDependencies: checkDependencies,
		}
	case soft.MetaCalculator:
		return &MetaCalculator{
			log:               logger,
			checkDependencies: checkDependencies,
		}
	case soft.AugmentedSemverCalculator:
		return &DefaultCalculator{
			log:               logger,
			checkDependencies: checkDependencies,
			scoreTable: segmentScoreTable{
				defaultMajorVersionScore * 2,
				defaultMajorVersionScore,
				defaultMinorVersionScore,
			},
		}
	case soft.SkipCalculator:
		return &SkipCalculator{}
	case soft.CandidateCountCalculator:
		return &CandidateCountCalculator{
			perCandidateScore: DefaultPerCandidateScore,
		}
	default:
		return &DefaultCalculator{
			log:               logger,
			checkDependencies: checkDependencies,
			scoreTable: segmentScoreTable{
				defaultMajorVersionScore,
				defaultMinorVersionScore,
				defaultPatchVersionScore,
			},
		}
	}
}

// Compute Obsolescence score by comparing current and last version
// and computes the obsolescence score
//
// x major versions late = m * 50
//
// y minor versions late = y * 5
//
// z patch versions late = z * 1
func (c *DefaultCalculator) CalculateObsolescenceScore(s *soft.Software) error {
	softwaresToCalculate := GetSoftwaresToCalculate(s, c.checkDependencies)
	c.log.Debug(fmt.Sprintf("Total of %d softwares to compute in order to compute software %s's total score", len(softwaresToCalculate), s.Name))
	topLevelScore := 0
	for _, software := range softwaresToCalculate {
		semver.Sort(software.VersionCandidates)
		// Retrieve semantic versions
		lv, err := goversion.NewSemver(software.VersionCandidates[0].Version)
		if err != nil {
			return fmt.Errorf("failed to parse latest version candidate's version %s using semver: %w", software.VersionCandidates[0].Version, err)
		}
		cv, err := goversion.NewSemver(software.Version.Version)
		if err != nil {
			return fmt.Errorf("failed to parse current version  %s using semver, %w", software.Version.Version, err)
		}
		latestVersion := lv.Segments()
		currentVersion := cv.Segments()
		c.log.Debug(fmt.Sprintf("Latest version found is %s, comparing with current version %s", lv, cv))

		// compute score
		for i := 0; i <= len(currentVersion)-1; i++ {
			diff := latestVersion[i] - currentVersion[i]
			if diff != 0 {
				topLevelScore += c.scoreTable[i] * diff
				software.CalculatedScore += c.scoreTable[i] * diff
				break
			}
		}
		s.CalculatedScore = topLevelScore
	}
	return nil
}

// Returns a list of softwares to compute a score for.
// If checkDependencies is true, include dependency softwares
// with a depth of 1 in the dependency tree
func GetSoftwaresToCalculate(s *soft.Software, checkDependencies bool) []*soft.Software {
	softwares := []*soft.Software{}

	// if a software is late (if it has at least one candidate) at top-level
	if len(s.VersionCandidates) > 0 {
		// then we calculate its score and don't care about the dependencies
		softwares = append(softwares, s)
	} else if checkDependencies {
		// No Version Candidates provided for top-level software.
		// then we calculate the score of its dependencies
		for _, dep := range s.Dependencies {
			if len(dep.VersionCandidates) < 1 {
				// No Version Candidates provided for dependency software
				// We skip the dependency
				continue
			} else {
				softwares = append(softwares, dep)
			}
		}
	}
	return softwares
}
