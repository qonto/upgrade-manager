package kubernetes

import (
	"context"
	"fmt"

	"github.com/qonto/upgrade-manager/internal/app/sources/helm/versions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ArgoCDApplication struct {
	Name                 string                   `json:"name"`
	DestinationNamespace string                   `json:"destination-namespace"`
	Project              string                   `json:"project"`
	Server               string                   `json:"server"`
	RepoURL              string                   `json:"repoURL"`
	ChartFilePath        string                   `json:"path"`
	Chart                string                   `json:"chart"`
	CurrentVersion       string                   `json:"version"`
	RepoBackendType      versions.RepoBackendType `json:"repobackendtype"` // Git or Helm
	GitRevision          string                   // master / dev / ...
}

var applicationGroup = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

// ListArgoApplications Retrieve helm-based argocd Applications from a kubernetes cluster's namespace'
func (c *Client) ListArgoApplications(ctx context.Context, namespace string, filters ...ArgoCDAppFilter) ([]*ArgoCDApplication, error) {
	var apps []*ArgoCDApplication

	rawApps, err := c.dynamicClient.Resource(applicationGroup).Namespace(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, rawApp := range rawApps.Items {
		app, err := rawToArgoApplication(rawApp.UnstructuredContent())
		if err != nil {
			c.logger.Info(fmt.Sprintf("Skipping app %s: not a properly deployed ArgoCD Helm App (error %s)", app.Name, err.Error()))
			continue
		}
		if len(filters) > 0 {
			for _, filter := range filters {
				if keep := filter(app); keep {
					apps = append(apps, app)
				}
			}
		} else {
			apps = append(apps, app)
		}
	}
	for _, app := range apps {
		c.logger.Debug(fmt.Sprintf("Tracking app %s with version %s, destination namespace: %s", app.Name, app.CurrentVersion, app.DestinationNamespace))
	}
	return apps, nil
}

func rawToArgoApplication(raw map[string]any) (*ArgoCDApplication, error) {
	name, found, err := unstructured.NestedString(raw, "metadata", "name")
	if err != nil || !found {
		return nil, err
	}
	newApp := &ArgoCDApplication{Name: name}

	_, isHelmApp, err := unstructured.NestedMap(raw, "spec", "source", "helm")
	if err != nil {
		return newApp, err
	}
	_, isDeployedApp, err := unstructured.NestedSlice(raw, "status", "history")
	if err != nil {
		return newApp, err
	}
	if !isHelmApp || !isDeployedApp {
		return newApp, fmt.Errorf("not a properly deployed Argo Helm application")
	}
	// These fields exist in all apps
	server, found, err := unstructured.NestedString(raw, "spec", "destination", "server")
	if err != nil || !found {
		return newApp, err
	}

	namespace, found, err := unstructured.NestedString(raw, "spec", "destination", "namespace")
	if err != nil || !found {
		return newApp, err
	}

	project, found, err := unstructured.NestedString(raw, "spec", "project")
	if err != nil || !found {
		return newApp, err
	}

	repoUrl, found, err := unstructured.NestedString(raw, "spec", "source", "repoURL")
	if err != nil || !found {
		return newApp, err
	}

	// At least one of these fields exist (either git repo with chart.yaml or helm repo with index.yaml)
	chartFilePath, chartFilePathFound, err := unstructured.NestedString(raw, "spec", "source", "path")
	if err != nil {
		return newApp, err
	}

	repoChart, helmChartFound, err := unstructured.NestedString(raw, "spec", "source", "chart")
	if err != nil {
		return newApp, err
	}
	if !chartFilePathFound && !helmChartFound {
		return newApp, fmt.Errorf("both spec.source.chart and spec.source.path were not defined in %s", name)
	}
	if chartFilePathFound {
		newApp.RepoBackendType = versions.GitRepo
		targetRevision, found, err := unstructured.NestedString(raw, "spec", "source", "targetRevision")
		if err != nil || !found {
			return newApp, err
		}
		newApp.GitRevision = targetRevision
	} else {
		newApp.RepoBackendType = versions.HelmRepo
	}

	currentRevision, found, err := unstructured.NestedString(raw, "status", "operationState", "syncResult", "revision")
	if err != nil || !found {
		return newApp, err
	}
	newApp.Server = server
	newApp.Project = project
	newApp.RepoURL = repoUrl
	newApp.ChartFilePath = chartFilePath
	newApp.Chart = repoChart
	newApp.CurrentVersion = currentRevision
	newApp.DestinationNamespace = namespace
	return newApp, nil
}
