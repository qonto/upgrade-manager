package semver

import (
	"errors"
	"fmt"
	"sort"

	goversion "github.com/hashicorp/go-version"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
)

var ErrorInSemverSortFunction = errors.New("cannot sort software semver versions")

func Sort(versions []software.Version) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w %v", ErrorInSemverSortFunction, r)
		}
	}()

	sort.Slice(versions, func(i, j int) bool {
		iVersion, err := goversion.NewSemver(versions[i].Version)
		if err != nil {
			panic(fmt.Errorf("cannot sort software %s versions: %w", versions[i].Name, err))
		}

		jVersion, err := goversion.NewSemver(versions[j].Version)
		if err != nil {
			panic(fmt.Errorf("cannot sort software %s versions: %w", versions[i].Name, err))
		}

		// Filtering out versions older than current version
		return iVersion.Core().Compare(jVersion.Core()) == 1
	})

	return nil
}

func ExtractFromString(rawString string) (string, error) {
	v, err := goversion.NewSemver(rawString)
	if err != nil {
		return "", err
	}
	return v.Core().String(), nil
}
