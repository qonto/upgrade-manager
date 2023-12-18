package versions

import (
	"os"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
)

// Load helm Chart.yaml file from local filesystem into a chart.Chart struct
func LoadChartFile(filePath string) (chart.Chart, error) {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return chart.Chart{}, err
	}
	var metadata chart.Metadata
	if err := yaml.Unmarshal(f, &metadata); err != nil {
		return chart.Chart{}, err
	}
	return chart.Chart{Metadata: &metadata}, nil
}
