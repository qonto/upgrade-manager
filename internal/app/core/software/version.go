package software

import "time"

type Versions []*Version

type Version struct {
	Name        string
	Version     string
	ReleaseDate time.Time
}

func (vs *Versions) Deduplicate() {
	allKeys := make(map[string]bool)
	list := Versions{}
	for _, item := range *vs {
		if _, value := allKeys[item.Version]; !value {
			allKeys[item.Version] = true
			list = append(list, item)
		}
	}
	*vs = list
}

func ToVersion(v string, optsfn ...func(*Version)) *Version {
	version := &Version{
		Version: v,
	}
	for _, fn := range optsfn {
		fn(version)
	}
	return version
}
