package deployments

import (
	"log/slog"
	"testing"

	"github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createDeployment(image string, deploymentName string, containerName string, calculator software.CalculatorType) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Annotations: map[string]string{
				calculatorAnnotation: string(calculator),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  containerName,
							Image: image,
						},
					},
				},
			},
		},
	}
}

// TODO: mock registry client
// for now it is performing actual calls to docker.io when running go test
func TestLoad(t *testing.T) {
	k8sMock := new(kubernetes.KubernetesClientMock)
	k8sMock.On("ListDeployments", mock.Anything).Return(
		&appsv1.DeploymentList{
			Items: []appsv1.Deployment{
				createDeployment("falcosecurity/falcosidekick:2.24.0", "sidekiq", "falcosidekick", software.ReleaseDateCalculator),
			},
		},
		nil)
	source, err := NewSource(slog.Default(), k8sMock, Config{})
	assert.NoError(t, err)
	softwares, err := source.Load()
	assert.NoError(t, err)
	assert.Equal(t, len(softwares), 1)
	assert.Equal(t, len(softwares[0].Dependencies), 1)
	assert.Equal(t, softwares[0].Dependencies[0].Name, "sidekiq-falcosidekick")
	assert.True(t, len(softwares[0].Dependencies[0].VersionCandidates) > 0)
	assert.True(t, softwares[0].Dependencies[0].VersionCandidates[0].ReleaseDate.IsZero())
	found := false
	for _, version := range softwares[0].Dependencies[0].VersionCandidates {
		if version.Version == "2.26.0" {
			found = true
			break
		}
	}
	assert.True(t, found, "Release 2.26.0 was not found")
}
