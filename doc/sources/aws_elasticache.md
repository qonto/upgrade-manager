# AWS Elasticache Source

### Description
The Elasticache source finds all the Elasticache clusters in the current AWS account with their current versions.

### New versions discovery
This source discovers new versions by calling the `GetCompatibleKafkaVersion` method with the default Go Elasticache client.

### Configuration

```yaml
sources:
    aws:
        elasticache:
            enabled: true
            request-timeout: 15s # (default: 15s)
```
Parameters:
- **request-timeout**: is the timeout duration when calling the AWS Elasticache API

### Obsolescence Score Calculation
The Elasticache source is using the [Semver Calculator](../calculators/semver_calculator.md)

