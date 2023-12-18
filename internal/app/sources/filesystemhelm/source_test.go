package filesystemhelm

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/stretchr/testify/mock"
)

func TestLoadSoftware(t *testing.T) {
	testChartsDir := "../../_testdata"
	expectedChartCount := 4
	cfg := Config{
		Enabled: true,
		Paths: []string{
			testChartsDir + "/test_chart/Chart.yaml",
			testChartsDir + "/test_chart2/Chart.yaml",
			testChartsDir + "/test_chart3/Chart.yaml",
			testChartsDir + "/test_chart4/Chart.yaml",
		},
	}
	indexFilePath := fmt.Sprintf("%s/index.yaml", testChartsDir)
	indexFile, err := os.ReadFile(indexFilePath)
	if err != nil {
		t.Fatal(err)
	}

	s3mock := new(aws.S3Mock)
	s3mock.On("GetObject", mock.Anything).Return(indexFile, nil)

	s, err := NewSource(cfg, slog.Default(), s3mock)
	if err != nil {
		t.Fatalf("Error with NewSource(): %s", err)
	}
	if len(s.Charts) != expectedChartCount {
		t.Fatalf("Wrong chart count returned from NewSource(). Expected : %d, got: %d", expectedChartCount, len(s.Charts))
	}
	softwares, err := s.Load()
	if err != nil {
		t.Fatalf("Error with Source.Load(): %s", err)
	}
	if len(softwares) != expectedChartCount {
		t.Fatalf("Wrong software count returned from Source.Load(). Expected : %d, got: %d", expectedChartCount, len(s.Charts))
	}
}
