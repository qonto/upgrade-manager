package versions

import (
	"testing"
)

func TestLoadChartFiles(t *testing.T) {
	testChartsDir := "../../../_testdata"
	paths := []string{
		testChartsDir + "/test_chart/Chart.yaml",
		testChartsDir + "/test_chart2/Chart.yaml",
		testChartsDir + "/test_chart3/Chart.yaml",
		testChartsDir + "/test_chart4/Chart.yaml",
	}
	for i, path := range paths {
		c, err := LoadChartFile(path)
		if err != nil {
			t.Error(err)
		}
		if c.Name() == "" || c.Metadata.Name == "" {
			t.Errorf("Chart %d: Not properly loaded helm chart file", i)
		}
	}
}
