package argohelm

import (
	"log/slog"
	"os"
	"regexp"
	"testing"

	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetGitCredentialSecretsFromNamespace(t *testing.T) {
	log := slog.Default()
	sampleKeyPath := "../../_testdata/fakeSampleKey"
	f, err := os.ReadFile(sampleKeyPath)
	if err != nil {
		t.Errorf("could not read sample private key at , %s", err)
	}
	k8sMock := new(kubernetes.KubernetesClientMock)
	k8sMock.On("ListSecrets", "argocd").Return(&v1.SecretList{
		Items: []v1.Secret{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "1-argocd-repo-secret-https",
					Namespace:   "argocd",
					Annotations: map[string]string{},
				},
				Data: map[string][]byte{
					"url":      []byte("https://repo1.git"),
					"username": []byte("myuser"),
					"password": []byte("mypassword"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "2-argocd-repo-secret-ssh",
					Namespace:   "argocd",
					Annotations: map[string]string{},
				},
				Data: map[string][]byte{
					"sshPrivateKey": f,
					"type":          []byte("ssh"),
					"url":           []byte("git@repo2:namespace/project.git"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "3-argocd-repo-secret-https-malformed", // missing username
					Namespace:   "argocd",
					Annotations: map[string]string{},
				},
				Data: map[string][]byte{
					"url":      []byte("https://repo1.git"),
					"password": []byte("mypassword"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "4-argocd-repo-secret-ssh-malformed", // malformed key
					Namespace:   "argocd",
					Annotations: map[string]string{},
				},
				Data: map[string][]byte{
					"sshPrivateKey": []byte("malformedkey"),
					"type":          []byte("ssh"),
					"url":           []byte("git@repo2:namespace/project.git"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "5-not-a-creds-secret",
					Namespace:   "argocd",
					Annotations: map[string]string{},
				},
			},
		},
	}, nil)
	r := regexp.MustCompile(".*-repo-.*")
	conns, err := getGitRepoConnections("argocd", r, k8sMock, log)
	if err != nil {
		t.Error(err)
	}
	if expectedConnCount := 2; len(conns) != expectedConnCount {
		t.Errorf("found wrong number of git connections. Expected %d, got: %d", expectedConnCount, len(conns))
	}
}
