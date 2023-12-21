# Semver Calculator

## Description
Using semantic versioning, this calculator compares the current and latest version, and compute the obsolescence score as follows:

+ x major versions late = `x * 50 point`

+ y minor versions late = `y * 5 points`

+ z patch versions late = `z * 1 point`

The calculator evaluates the version segments one by one in the following order: `major>minor>patch`. As soon as it finds a difference between the current and the latest version, it computes the score for the segment and skips the following ones.

Sources using this Calculator:
- [aws_eks](../sources/aws_eks.md) (for eks addons versions)
- [aws_elasticache](../sources/aws_elasticache.md)
- [aws_rds](../sources/aws_rds.md)
- [aws_msk](../sources/aws_msk.md)
- [argoCDHelm](../sources/argoCDHelm.md)

## Augmented Semver Calculator
Some softwares following semantic versioning have a major version that almost never changes. AWS EKS (kubernetes) is a good example as we don't know if we'll ever going to see a major version 2.0 in the future. In the case of Kubernetes, minor versions can actually be considered major versions as they introduce significant (sometimes breaking) changes and as most major cloud and software providers generally support only a few minor versions. From that fact, we decided to introduce the Augmented Semver Calculator.

The exact score table could be made configurable per source in the future.

The `AugmentedSemverCalculator` is working the same way as the Semver Calculator, but it adds more points:

+ x major versions late = `x * 100 point`

+ y minor versions late = `y * 50 points`

+ z patch versions late = `z * 5 point`

The `AugmentedSemverCalculator` is currently only used by the EKS source.

Sources using this Calculator:
- [aws_eks](../sources/aws_eks.md) (for k8s engine versions)

## Example
An ArgoHelm (ArgoCD Application target a Helm configuration) software has a current Chart version of `1.2.1` and the latest eligible version is `1.6.3`.

1. The calculator compares the major versions: `1` and `1`. They are the same so it does not add any points
2. The calculator compares the minor versions: `2` and `6`. Since `6` is higher than `2`, the calculator adds `(6-2)*5 = 4*5 = 20 points`
3. The calcualtor does not compare the patch versions because it has already found a difference in a previous version segment.

