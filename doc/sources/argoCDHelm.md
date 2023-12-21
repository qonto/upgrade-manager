# AWS ArgoCDHelm Source

### Description
The ArgoCDHelm source finds all the [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) Applications targeting a Helm configuration. A Helm configuration is an application with a populated `spec.source.helm` field of the Application object.
It does so by listing Applications objects in the namespace defined in `config.sources.argocdhelm.argocd-namespace`.
When it finds an application, it clones the source repository and parses the `Chart.yaml` file.
If the `Chart.yaml` file references Chart dependencies the obsolescence scores of the dependencies will be added to the software only if the top-level chart is already up-to-date.

### New versions discovery
This source discovers new versions by fetching the `index.yaml` file available at the root of the helm repository. The `index.yaml` file lists all versions for a given chart.

### Git Repository Connection
In order to clone private git repositories, `upgrade-manager` requires credentials. 
1. At startup, upgrade-manager lists `secrets` in its namespace or whatever is defined by `git-credentials-secrets-namespace`.
2. Then, it loads credentials from all secrets matching this regular expression: `.*-repo.*` or whatever is defined by `git-credentials-secrets-pattern`
The HTTPS connection secrets should have the following keys:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: private-repo-creds-https
  namespace: upgrade-manager
data:
  password: password-in-base64
  url: git-url-in-base64
  username: username-in-base64
```

After loading the secrets, upgrade-manager will use them to connect software's repositories using a prefix-based approach. 
For example: 
- Let's say we have two secrets:
Secret 1 is providing credentials to connect to the git url `https://foo.bar.com`
Secret 2 is providing credentials to connect to the git url `https://foo.bar.com/team1`
- Now, let's say we have an `app1` ArgoCD Helm application deployed with its source repository available at `https://foo.bar.com/team1/app1`
    - upgrade-manager will use Secret 2's credentials to connect to the instance

For test purposes, it is recommended to start with a single secret pointing at the root of your git repository. The credentials need readonly access to the desired repositories.

### Custom filtering logic to assess a version's eligibility
In order to customize which Applications upgrade-manager should keep and which versions upgrade-manager selects as the best version candidates for an ArgoCD Helm software, upgrade-manager uses `filters` (see the detailed list below).

### Configuration

```yaml
sources:
  argocdHelm:
    - enabled: true
      argocd-namespace: argocd # namespace where the argocd application object is deployed
      git-credentials-secrets-namespace: upgrade-manager # namespace where secrets containing git credentials are deployed
      git-credentials-secrets-pattern: ".*-repo-.*" # regex to filter which secrets to fetch
      filters:
        semver-versions:
          remove-pre-release: true
          remove-first-major-version: true
        recent-versions:
          days: 21 # number of days since the version was released
        destination-namespace: # namespace where the app resources will be deployed
          include: []
          # ["kyverno"] # regular expression, if include is empty, it includes all namespaces
          exclude: []
          # ["default", "temporary-*", "feature-*", "kube-system"] # regular expression if exclude is empty, it does not exclude any namespace
 ```
Parameters:
- **argocd-namespace**: namespace where the ArgoCD Application objects are deployed.
- **git-credentials-secrets-namespace**: namespace where secrets containing git credentials are deployed.
- **git-credentials-secrets-pattern**: regex to filter which secrets to load git credentials from.
- **filters.semver-versions.remove-pre-release**: whether pre-release versions should be considered as eligible version candidates or skipped.
- **filters.semver-versions.remove-first-major-version**: whether versions X.0.0 should be considered as eligible candidates or skipped.
- **filters.recent-versions.days**: the minimum age of a version before it should be considered as an eligible version candidate.
- **filters.destination-namespace.include**: list of destination namespaces (corresponding to the destination namespace in an ArgoCD application spec) to compute a score/discover new targets for. If `include:` is empty, it includes all namespaces. Wildcard patterns (ex: `feat-*`) can be used.
- **filters.destination-namespace.exclude**: list of destination namespaces (corresponding to the destination namespace in an ArgoCD application spec) to avoid computing a score/discover new targets for. If `exclude:` is empty, it does not exclude any namespace. Wildcard patterns (ex: `feat-*`) can be used.

### Known limitations
Some limitations are known and will be tackled in future updates:
- add support for ArgoCD Helm repositories aliases in the form of "@myalias or alias:myalias"
- add support for dynamic revision tracking strategies (^0.1, 1.2.* etc.) by failing-over using the helm.sh/chart label

### Obsolescence score calculation
The ArgoCD Helm source is using the [Semver Calculator](../calculators/semver_calculator.md)