package semver

import (
	"sort"

	goversion "github.com/hashicorp/go-version"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
)

func Sort(versions []software.Version) {
	sort.Slice(versions, func(i, j int) bool {
		iVersion, _ := goversion.NewSemver(versions[i].Version)
		jVersion, _ := goversion.NewSemver(versions[j].Version)
		// Filtering out versions older than current version
		return iVersion.Core().Compare(jVersion.Core()) == 1
	})
}

func ExtractFromString(rawString string) (string, error) {
	v, err := goversion.NewSemver(rawString)
	if err != nil {
		return "", err
	}
	return v.Core().String(), nil
}
