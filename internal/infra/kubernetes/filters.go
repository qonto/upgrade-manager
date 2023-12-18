package kubernetes

import (
	"regexp"
)

type ArgoCDAppFilter func(*ArgoCDApplication) bool

type FiltersOptions struct {
	// Destination namespace of Argocd Applications
	DestinationNamespaceFilterOptions `yaml:"destination-namespace"`
	// TODO:
	// filter on app name, app labels, transform with annotations etc...
}

type DestinationNamespaceFilterOptions struct {
	NamespaceFilterOptions `yaml:",inline"`
}

type NamespaceFilterOptions struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

func NewDestinationNamespaceFilter(opts FiltersOptions) (ArgoCDAppFilter, error) {
	includeExpr, excludeExpr, err := opts.NamespaceFilterOptions.Compile()
	if err != nil {
		return nil, err
	}
	return func(app *ArgoCDApplication) bool {
		return FilterNamespace(app.DestinationNamespace, includeExpr, excludeExpr)
	}, nil
}

// Returns a converted-to-*regexp.Regexp version of NamespaceFilterOptions.Include and NamespaceFilterOptions.Exclude
func (ds NamespaceFilterOptions) Compile() ([]*regexp.Regexp, []*regexp.Regexp, error) {
	excludeExpr := make([]*regexp.Regexp, 0, len(ds.Exclude))
	includeExpr := make([]*regexp.Regexp, 0, len(ds.Include))

	for _, ns := range ds.Exclude {
		expr, err := regexp.Compile(ns)
		if err != nil {
			return nil, nil, err
		}
		excludeExpr = append(excludeExpr, expr)
	}
	for _, ns := range ds.Include {
		expr, err := regexp.Compile(ns)
		if err != nil {
			return nil, nil, err
		}
		includeExpr = append(includeExpr, expr)
	}
	return includeExpr, excludeExpr, nil
}

func FilterNamespace(namespace string, includeExpr []*regexp.Regexp, excludeExpr []*regexp.Regexp) bool {
	for _, expr := range excludeExpr {
		if expr.MatchString(namespace) {
			return false
		}
	}
	for _, expr := range includeExpr {
		if expr.String() != "" {
			if expr.MatchString(namespace) {
				return true
			}
		}
	}
	return len(includeExpr) == 0
}
