# Date Calculator

## Description
For each day since the day the latest eligible version candidate was released, the software will receive an extra `5` obsolescence points.

Sources using this Calculator:
- [Deployments](../sources/deployments.md)

### Example
A deployment is using an image `foo:1.1`
The latest eligible version is `foo:1.4`, it was released `30` days after `foo:1.1`

In that case, the deployment software will be given an obsolescence of `30*5=60 points`