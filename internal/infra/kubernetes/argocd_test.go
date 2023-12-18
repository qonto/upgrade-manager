package kubernetes

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func TestRawToArgoApplication(t *testing.T) {
	type testCase struct {
		argoApp         map[string]any
		expectedSuccess bool
	}
	testCases := []testCase{
		{
			expectedSuccess: true,
			argoApp: map[string]any{
				"metadata": map[string]any{
					"name": "normalapp",
				}, "spec": map[string]any{
					"source": map[string]any{
						"helm": map[string]any{
							"releaseName": "alertmanager-webhook-logger",
						},
						"repoURL":        "https://github.qonto.co/devops/kubernetes-resources/alertmanager-webhook-logger.git",
						"path":           "data-staging",
						"targetRevision": "master",
					},
					"project": "monitoring",
					"destination": map[string]any{
						"server":    "https://kubernetes.default.svc",
						"namespace": "monitoring",
					},
				},
				"status": map[string]any{
					"history": []any{},
					"operationState": map[string]any{
						"syncResult": map[string]any{
							"revision": "1.0.0",
						},
					},
				},
			},
		},
		{
			expectedSuccess: false,
			argoApp: map[string]any{
				"metadata": map[string]any{
					"name": "nostatusapp",
				}, "spec": map[string]any{
					"source": map[string]any{
						"helm": map[string]any{
							"releaseName": "alertmanager-webhook-logger",
						},
						"repoURL":        "https://github.qonto.co/devops/kubernetes-resources/alertmanager-webhook-logger.git",
						"path":           "data-staging",
						"targetRevision": "master",
					},
					"project": "monitoring",
					"destination": map[string]any{
						"server":    "https://kubernetes.default.svc",
						"namespace": "monitoring",
					},
				},
			},
		},
		{
			expectedSuccess: false,
			argoApp: map[string]any{
				"metadata": map[string]any{
					"name": "notahelmapp",
				}, "spec": map[string]any{
					"source": map[string]any{
						"repoURL":        "https://github.qonto.co/devops/kubernetes-resources/alertmanager-webhook-logger.git",
						"path":           "data-staging",
						"targetRevision": "master",
					},
					"project": "monitoring",
					"destination": map[string]any{
						"server":    "https://kubernetes.default.svc",
						"namespace": "monitoring",
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		_, err := rawToArgoApplication(tc.argoApp)
		if err != nil && tc.expectedSuccess {
			t.Fatalf("Case %d: failed to convert raw app but it should have succeeded, %s", i+1, err)
		}
		if err == nil && !tc.expectedSuccess {
			t.Fatalf("Case %d: successfully converted raw app but it should not have succeeded, %s", i+1, err)
		}
	}
}

func TestListArgoApplications(t *testing.T) {
	log := zap.NewExample()
	testCases := []map[string]any{
		{
			"kind":       "Application",
			"apiVersion": "argoproj.io/v1alpha1",
			"metadata": map[string]any{
				"name":      "normalapp",
				"namespace": "argocd",
			}, "spec": map[string]any{
				"source": map[string]any{
					"helm": map[string]any{
						"releaseName": "alertmanager-webhook-logger",
					},
					"repoURL":        "https://github.qonto.co/devops/kubernetes-resources/alertmanager-webhook-logger.git",
					"path":           "data-staging",
					"targetRevision": "master",
				},
				"project": "monitoring",
				"destination": map[string]any{
					"server":    "https://kubernetes.default.svc",
					"namespace": "monitoring",
				},
			},
			"status": map[string]any{
				"history": []any{},
				"operationState": map[string]any{
					"syncResult": map[string]any{
						"revision": "1.0.0",
					},
				},
			},
		},
		{
			"kind":       "Application",
			"apiVersion": "argoproj.io/v1alpha1",
			"metadata": map[string]any{
				"name":      "normalapp2",
				"namespace": "argocd",
			}, "spec": map[string]any{
				"source": map[string]any{
					"helm": map[string]any{
						"releaseName": "alertmanager-webhook-logger",
					},
					"repoURL":        "https://github.qonto.co/devops/kubernetes-resources/alertmanager-webhook-logger.git",
					"path":           "data-staging",
					"targetRevision": "master",
				},
				"project": "monitoring",
				"destination": map[string]any{
					"server":    "https://kubernetes.default.svc",
					"namespace": "monitoring",
				},
			},
			"status": map[string]any{
				"history": []any{},
				"operationState": map[string]any{
					"syncResult": map[string]any{
						"revision": "2.0.0",
					},
				},
			},
		},
		{
			"kind":       "Application",
			"apiVersion": "argoproj.io/v1alpha1",
			"metadata": map[string]any{
				"name":      "nohistoryapp",
				"namespace": "argocd",
			}, "spec": map[string]any{
				"source": map[string]any{
					"helm": map[string]any{
						"releaseName": "alertmanager-webhook-logger",
					},
					"repoURL":        "https://github.qonto.co/devops/kubernetes-resources/alertmanager-webhook-logger.git",
					"path":           "data-staging",
					"targetRevision": "master",
				},
				"project": "monitoring",
				"destination": map[string]any{
					"server":    "https://kubernetes.default.svc",
					"namespace": "monitoring",
				},
			},
			"status": map[string]any{
				"operationState": map[string]any{
					"syncResult": map[string]any{
						"revision": "2.0.0",
					},
				},
			},
		},
	}
	client := fake.NewSimpleDynamicClient(&runtime.Scheme{},
		&unstructured.Unstructured{Object: testCases[0]},
		&unstructured.Unstructured{Object: testCases[1]},
		&unstructured.Unstructured{Object: testCases[2]},
	)
	s := Client{dynamicClient: client, logger: log}
	apps, err := s.ListArgoApplications(context.Background(), "argocd")
	if err != nil {
		t.Error(err)
	}
	if expectedAppCount := 2; len(apps) != expectedAppCount {
		t.Errorf("unexpected number of Argocd applications found. Expected: %d, got: %d", expectedAppCount, len(apps))
	}
}
