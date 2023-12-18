package kubernetes

import (
	"context"

	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type KubernetesClientMock struct {
	mock.Mock
}

func (m *KubernetesClientMock) ListArgoApplications(ctx context.Context, namespace string, filters ...ArgoCDAppFilter) ([]*ArgoCDApplication, error) {
	return nil, nil
}

func (m *KubernetesClientMock) ListDeployments(ctx context.Context, request ListRequest) (*appsv1.DeploymentList, error) {
	args := m.Called(request.Namespace)
	return args.Get(0).(*appsv1.DeploymentList), args.Error(1) //nolint
}

func (m *KubernetesClientMock) ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error) {
	args := m.Called(namespace)
	secrets := args.Get(0).(*v1.SecretList) //nolint
	return secrets, args.Error(1)
}
