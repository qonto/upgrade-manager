package eks

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/filters"
	awsInfra "github.com/qonto/upgrade-manager/internal/infra/aws"
	"log/slog"
)

func TestLoad(t *testing.T) {
	currentEksVersion := "1.23"
	latestEksVersion := "1.24"
	currentAddonVersion := "1.3.2"
	latestAddonVersion := "1.5.0"
	cls := []struct {
		name               string
		k8sVersion         string
		k8sExpectedVersion string
		expectedCalculator software.CalculatorType
		addons             []struct {
			name                  string
			version               string
			versions              []string
			expectedLatestVersion string
			expectedCalculator    software.CalculatorType
		}
	}{
		{
			name:               "cluster1",
			k8sVersion:         currentEksVersion,
			k8sExpectedVersion: latestEksVersion,
			addons: []struct {
				name                  string
				version               string
				versions              []string
				expectedLatestVersion string
				expectedCalculator    software.CalculatorType
			}{
				{
					name:                  "ebs-csi-driver",
					version:               fmt.Sprintf("%s-eksbuild.v2", currentAddonVersion),
					expectedLatestVersion: latestAddonVersion,
					expectedCalculator:    software.MetaCalculator,
				},
			},
		},
	}
	mockApi := new(awsInfra.EksMock)
	clusterNames := []string{}

	for _, cluster := range cls {
		clusterNames = append(clusterNames, cluster.name)
	}

	mockApi.On("ListClusters").Return(
		&eks.ListClustersOutput{
			Clusters: clusterNames,
		},
	)
	mockApi.On("ListAddons", eks.ListAddonsInput{ClusterName: &cls[0].name}).Return(
		&eks.ListAddonsOutput{
			Addons: []string{cls[0].addons[0].name},
		},
	)
	mockApi.On("DescribeAddonVersions").Return(
		&eks.DescribeAddonVersionsOutput{
			Addons: []types.AddonInfo{
				{
					AddonVersions: []types.AddonVersionInfo{
						{
							Compatibilities: []types.Compatibility{{ClusterVersion: &latestEksVersion}},
							AddonVersion:    aws.String(fmt.Sprintf("%s-eksbuild.v4", cls[0].addons[0].expectedLatestVersion)),
						},
						{
							Compatibilities: []types.Compatibility{{ClusterVersion: &latestEksVersion}},
							AddonVersion:    aws.String(fmt.Sprintf("%s-eksbuild.v3", cls[0].addons[0].expectedLatestVersion)),
						},
					},
					AddonName: aws.String("ebs-csi-driver"),
				},
			},
		})
	mockApi.On("DescribeCluster", eks.DescribeClusterInput{Name: &cls[0].name}).Return(
		&eks.DescribeClusterOutput{
			Cluster: &types.Cluster{
				Version: &cls[0].k8sVersion,
			},
		},
	)
	mockApi.On("DescribeAddon", eks.DescribeAddonInput{
		AddonName:   &cls[0].addons[0].name,
		ClusterName: &cls[0].name,
	}).Return(
		&eks.DescribeAddonOutput{
			Addon: &types.Addon{
				AddonName:    &cls[0].addons[0].name,
				AddonVersion: &cls[0].addons[0].version,
			},
		},
	)
	src, err := NewSource(mockApi, slog.Default(), &Config{
		Enabled: true,
		Filters: filters.Config{
			SemverVersions: &filters.SemverVersionsConfig{
				RemovePreRelease: true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	softs, err := src.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(softs) != 1 {
		t.Error(fmt.Errorf("%s", "unexpected number of softwares"))
	}
	if softs[0].Calculator != software.MetaCalculator {
		t.Errorf("wrong calculator type for soft %s", softs[0].Name)
	}
	for _, dep := range softs[0].Dependencies {
		if dep.Name == "ebs-csi-driver" && dep.Calculator != software.SemverCalculator && dep.VersionCandidates[0].Version != latestAddonVersion {
			t.Errorf("wrong result for dep %s", dep.Name)
		}
		if dep.Name == "k8s" && dep.Calculator != software.AugmentedSemverCalculator && dep.VersionCandidates[0].Version != latestEksVersion {
			t.Errorf("wrong result for dep %s", dep.Name)
		}
	}
}
