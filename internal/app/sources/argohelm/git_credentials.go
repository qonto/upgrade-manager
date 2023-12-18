package argohelm

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/qonto/upgrade-manager/internal/app/sources/utils/gitutils"
	k8sClient "github.com/qonto/upgrade-manager/internal/infra/kubernetes"
	"log/slog"
)

// Retrieve all git credentials in the namespace which have "repo" in their name
//
// We assume the secrets follow the argocd schema for secret.Data, meaning that
// git http secrets have the following data keys: "type: git", "password; xxx",
// "url: yyyxxx", "username: zzz"
//
// ssh secrets have "type: ssh", "sshPrivateKey; xxx", "url: yyyxxx"
func getGitRepoConnections(namespace string, r *regexp.Regexp, client k8sClient.KubernetesClient, log *slog.Logger) ([]*gitutils.RepoConnection, error) {
	var connections []*gitutils.RepoConnection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	secrets, err := client.ListSecrets(ctx, namespace)
	if err != nil {
		return nil, err
	}
	log.Debug(fmt.Sprintf("Found %d secrets in namespace %s", len(secrets.Items), namespace))

	for _, secret := range secrets.Items {
		isRepoCredSecret := r.MatchString(secret.Name)
		// TODO: handle case where ssh secret has no url to add the key to the local ssh-agent
		// (or just implement trying all keys when trying a repo clone via ssh before using default auth?)
		if isRepoCredSecret && secret.Data["url"] != nil {
			repo := &gitutils.RepoConnection{}
			repo.Url = string(secret.Data["url"])
			switch {
			case secret.Data["username"] != nil && secret.Data["password"] != nil:
				// if http connection credentials return http auth type with username/password
				repo.Private = true
				repo.Auth = &http.BasicAuth{Username: string(secret.Data["username"]), Password: string(secret.Data["password"])}
				log.Debug(fmt.Sprintf("Adding connection of type https to git url %s from secret %s", repo.Url, secret.Name))
				connections = append(connections, repo)
			case secret.Data["sshPrivateKey"] != nil && string(secret.Data["type"]) == "ssh":
				// if ssh connection credentials return ssh auth type with the key
				keys, err := ssh.NewPublicKeys("upgrade-manager", secret.Data["sshPrivateKey"], "")
				if err != nil {
					log.Warn(fmt.Sprintf("skipping secret %s: could not load key. %s", secret.Name, err))
					continue
				}
				repo.Auth = keys
				repo.Private = true
				connections = append(connections, repo)
			default:
				log.Debug(fmt.Sprintf("Skipping repo secret %s: not a proper ssh or https git repo. isRepoCredsSecret :%s, url:%s", secret.Name, strconv.FormatBool(isRepoCredSecret), string(secret.Data["url"])))
				continue
			}
		}
	}
	return connections, err
}
