package kubernetes

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ListRequest struct {
	Namespace     string
	LabelSelector map[string]string
}

func (c *Client) ListDeployments(ctx context.Context, request ListRequest) (*v1.DeploymentList, error) {
	labelSelector := metav1.LabelSelector{MatchLabels: request.LabelSelector}

	return c.kubernetesClient.AppsV1().Deployments(request.Namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})
}
