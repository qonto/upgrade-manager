# AWS FileSystemHelm Source

### Description
The FileSystemHelm source parses all Chart.yaml files provided.
If the `Chart.yaml` file references Chart dependencies the obsolescence scores of the dependencies will be added to the software only if the top-level chart is already up-to-date.

### New versions discovery
This source discovers new versions by fetching the `index.yaml` file available at the root of the helm repository. The `index.yaml` file lists all versions for a given chart.

### Custom filtering logic to assess a version's eligibility
In order to customize which Applications upgrade-manager should keep and which versions upgrade-manager selects as the best version candidates for an ArgoCD Helm software, upgrade-manager uses `filters`.


### Configuration

```yaml
sources:
  filesystemHelm:
    - enabled: true
      paths:
        - "./test_chart/Chart.yaml"
        - "./test_chart2/Chart.yaml"
      filters:
        semver-versions:
          remove-pre-release: true
          remove-first-major-version: true
        recent-versions:
          days: 21
 ```
Parameters:
- **filters.semver-versions.remove-pre-release**: whether pre-release versions should be considered as eligible version candidates or skipped.
- **filters.semver-versions.remove-first-major-version**: whether versions X.0.0 should be considered as eligible candidates or skipped.
- **filters.recent-versions.days**: the minimum age of a version before it should be considered as an eligible version candidate.
- **filters.destination-namespace.include**: list of destination namespaces (corresponding to the destination namespace in an ArgoCD application spec) to compute a score/discover new targets for. If `include:` is empty, it includes all namespaces. Wildcard patterns (ex: `feat-*`) can be used.
- **filters.destination-namespace.exclude**: list of destination namespaces (corresponding to the destination namespace in an ArgoCD application spec) to avoid computing a score/discover new targets for. If `exclude:` is empty, it does not exclude any namespace. Wildcard patterns (ex: `feat-*`) can be used.


### Obsolescence Score Calculation
The ArgoCD Helm source is using the [Semver Calculator](../calculators/semver_calculator.md)