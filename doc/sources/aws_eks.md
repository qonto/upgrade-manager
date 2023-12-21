# AWS EKS Source

### Description
The EKS source finds all the EKS clusters in the current AWS account with the current versions for their k8s engine and the different EKS addons deployed (vpc-cni, kube-proxy etc.).

### New versions discovery
This source discovers new versions by calling the `DescribeAddonVersions` method with the default Go EKS client.

### Configuration

```yaml
sources:
    aws:
        eks:
            enabled: true
            request-timeout: 15s # (default: 15s)
```
Parameters:
- **request-timeout**: is the timeout duration when calling the AWS EKS API

### Obsolescence Score Calculation
The EKS source is using the [Semver Calculator](../calculators/semver_calculator.md) for `addons` and [Augmented Semver Calculator](../calculators/semver_calculator.md#augmented-semver-calculator) for the `k8s-engine`.