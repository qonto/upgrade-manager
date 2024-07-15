package msk

import (
	"log/slog"
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/core/software"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/qonto/upgrade-manager/internal/app/sources/utils"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/stretchr/testify/mock"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		name                      string
		initFunc                  func(*aws.MockMSKApi)
		expectedError             bool
		expectedClusterCount      int
		expectedVersionCandidates []string
	}{
		{
			name: "happy path",
			initFunc: func(api *aws.MockMSKApi) {
				api.On("ListClustersV2", mock.Anything).Return(
					&kafka.ListClustersV2Output{
						ClusterInfoList: []types.Cluster{
							{
								ClusterName: utils.Ptr("mycluster"),
								ClusterArn:  utils.Ptr("arn:myclusterarn"),
								Provisioned: &types.Provisioned{
									CurrentBrokerSoftwareInfo: &types.BrokerSoftwareInfo{
										KafkaVersion: utils.Ptr("2.0.0"),
									},
								},
							},
						},
					})

				api.On("GetCompatibleKafkaVersions", mock.Anything).Return(
					&kafka.GetCompatibleKafkaVersionsOutput{
						CompatibleKafkaVersions: []types.CompatibleKafkaVersion{
							{
								TargetVersions: []string{
									"2.2.3",
									"2.3.4",
								},
							},
						},
					})
			},
			expectedError:        false,
			expectedClusterCount: 1,
			expectedVersionCandidates: []string{
				"2.2.3",
				"2.3.4",
			},
		},
		{
			name: "msk special versions",
			initFunc: func(api *aws.MockMSKApi) {
				api.On("ListClustersV2", mock.Anything).Return(
					&kafka.ListClustersV2Output{
						ClusterInfoList: []types.Cluster{
							{
								ClusterName: utils.Ptr("mycluster"),
								ClusterArn:  utils.Ptr("arn:myclusterarn"),
								Provisioned: &types.Provisioned{
									CurrentBrokerSoftwareInfo: &types.BrokerSoftwareInfo{
										KafkaVersion: utils.Ptr("2.0.0"),
									},
								},
							},
						},
					})

				api.On("GetCompatibleKafkaVersions", mock.Anything).Return(
					&kafka.GetCompatibleKafkaVersionsOutput{
						CompatibleKafkaVersions: []types.CompatibleKafkaVersion{
							{
								TargetVersions: []string{
									"2.2.3",
									"2.3.4.tiered",
									"3.7.x",
								},
							},
						},
					})
			},
			expectedError:        false,
			expectedClusterCount: 1,
			expectedVersionCandidates: []string{
				"2.2.3",
				"2.3.4",
				"3.7.0",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := new(aws.MockMSKApi)
			tc.initFunc(api)

			source, err := NewSource(api, slog.Default(), &Config{})
			if err != nil {
				t.Error(err)
			}
			softwares, err := source.Load()
			if err != nil {
				t.Error(err)
			}

			if len(softwares) != tc.expectedClusterCount {
				t.Errorf("expected %d cluster", tc.expectedClusterCount)
			}

			if len(softwares[0].VersionCandidates) != len(tc.expectedVersionCandidates) {
				t.Errorf("expected %d version candidates, got %d", len(tc.expectedVersionCandidates), len(softwares[0].VersionCandidates))
			}

			for _, expectedCandidate := range tc.expectedVersionCandidates {
				if !contains(softwares[0].VersionCandidates, expectedCandidate) {
					t.Errorf("does not find version %s in result", expectedCandidate)
				}
			}
		})
	}
}

func contains(slice []software.Version, value string) bool {
	for _, v := range slice {
		if v.Version == value {
			return true
		}
	}
	return false
}
