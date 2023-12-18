package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithDestinationNamespaceFilter(t *testing.T) {
	testCases := []struct {
		description   string
		opts          FiltersOptions
		expectedCount int
		apps          []*ArgoCDApplication
	}{
		{
			description:   "Include some and remove some",
			expectedCount: 1,
			opts: FiltersOptions{
				DestinationNamespaceFilterOptions: DestinationNamespaceFilterOptions{
					NamespaceFilterOptions: NamespaceFilterOptions{
						Include: []string{"kyverno"},
						Exclude: []string{"master", "deps-*"},
					},
				},
			},
			apps: []*ArgoCDApplication{
				{Name: "App1", DestinationNamespace: "kyverno"},
				{Name: "App2", DestinationNamespace: "master"},
				{Name: "App3", DestinationNamespace: "deps-123"},
				{Name: "App4", DestinationNamespace: "deps-345"},
			},
		},
		{
			description:   "Include none and remove one",
			expectedCount: 3,
			opts: FiltersOptions{
				DestinationNamespaceFilterOptions: DestinationNamespaceFilterOptions{
					NamespaceFilterOptions: NamespaceFilterOptions{
						Include: nil,
						Exclude: []string{"master", "ns2"},
					},
				},
			},
			apps: []*ArgoCDApplication{
				{Name: "App1", DestinationNamespace: "kyverno"},
				{Name: "App2", DestinationNamespace: "master"},
				{Name: "App3", DestinationNamespace: "deps-123"},
				{Name: "App4", DestinationNamespace: "deps-345"},
			},
		},
		{
			description:   "Include empty string, remove some",
			expectedCount: 0,
			opts: FiltersOptions{
				DestinationNamespaceFilterOptions: DestinationNamespaceFilterOptions{
					NamespaceFilterOptions: NamespaceFilterOptions{
						Include: []string{""},
						Exclude: []string{"master", "deps-*"},
					},
				},
			},
			apps: []*ArgoCDApplication{
				{Name: "App1", DestinationNamespace: "kyverno"},
				{Name: "App2", DestinationNamespace: "master"},
				{Name: "App3", DestinationNamespace: "deps-123"},
				{Name: "App4", DestinationNamespace: "deps-345"},
			},
		},
		{
			description:   "Include none and remove none",
			expectedCount: 4,
			opts: FiltersOptions{
				DestinationNamespaceFilterOptions: DestinationNamespaceFilterOptions{},
			},
			apps: []*ArgoCDApplication{
				{Name: "App1", DestinationNamespace: "kyverno"},
				{Name: "App2", DestinationNamespace: "master"},
				{Name: "App3", DestinationNamespace: "deps-123"},
				{Name: "App4", DestinationNamespace: "deps-345"},
			},
		},
		{
			description:   "Include some and remove same",
			expectedCount: 0,
			opts: FiltersOptions{
				DestinationNamespaceFilterOptions: DestinationNamespaceFilterOptions{
					NamespaceFilterOptions: NamespaceFilterOptions{
						Include: []string{"master"},
						Exclude: []string{"master"},
					},
				},
			},
			apps: []*ArgoCDApplication{
				{Name: "App1", DestinationNamespace: "kyverno"},
				{Name: "App2", DestinationNamespace: "master"},
				{Name: "App3", DestinationNamespace: "deps-123"},
				{Name: "App4", DestinationNamespace: "deps-345"},
			},
		},
		{
			description:   "Include some and remove none",
			expectedCount: 1,
			opts: FiltersOptions{
				DestinationNamespaceFilterOptions: DestinationNamespaceFilterOptions{
					NamespaceFilterOptions: NamespaceFilterOptions{
						Include: []string{"master"},
					},
				},
			},
			apps: []*ArgoCDApplication{
				{Name: "App1", DestinationNamespace: "kyverno"},
				{Name: "App2", DestinationNamespace: "master"},
				{Name: "App3", DestinationNamespace: "deps-123"},
				{Name: "App4", DestinationNamespace: "deps-345"},
			},
		},
	}

	// FilterNamespace(opts)
	for _, tc := range testCases {
		var argoFilters []ArgoCDAppFilter
		filter, err := NewDestinationNamespaceFilter(tc.opts)
		if err != nil {
			t.Error(err)
		}
		argoFilters = append(argoFilters, filter)
		kept := []*ArgoCDApplication{}
		for _, app := range tc.apps {
			for _, filter := range argoFilters {
				if keep := filter(app); keep {
					kept = append(kept, app)
				}
			}
		}
		assert.Equal(t, tc.expectedCount, len(kept))
	}
}
