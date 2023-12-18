package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	dynamicClient    dynamic.Interface
	kubernetesClient kubernetes.Interface
	logger           *slog.Logger
}

type KubernetesClient interface {
	ListArgoApplications(ctx context.Context, namespace string, filters ...ArgoCDAppFilter) ([]*ArgoCDApplication, error)
	ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error)
	ListDeployments(ctx context.Context, request ListRequest) (*appsv1.DeploymentList, error)
}

func NewClient(logger *slog.Logger) (*Client, error) {
	var config *rest.Config
	var err error

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if _, err := os.Stat(kubeconfig); err != nil {
		kubeconfig = ""
	}
	if kubeconfig == "" {
		if os.Getenv("KUBERNETES_SERVICE_HOST") == "" || os.Getenv("KUBERNETES_SERVICE_PORT") == "" {
			return nil, fmt.Errorf("kubernetes environment variables not defined")
		}
		config, err = rest.InClusterConfig()

		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		logger:           logger,
		dynamicClient:    dynamicClient,
		kubernetesClient: client,
	}, nil
}
