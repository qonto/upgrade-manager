package msk

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/qonto/upgrade-manager/internal/app/sources/utils"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestLoad(t *testing.T) {
	api := new(aws.MockMSKApi)
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
					},
				},
			},
		})
	source, err := NewSource(api, zap.NewExample(), &Config{})
	if err != nil {
		t.Error(err)
	}
	_, err = source.Load()
	if err != nil {
		t.Error(err)
	}
}
