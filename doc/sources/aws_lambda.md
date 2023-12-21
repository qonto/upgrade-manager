# AWS Lambda Source

### Description
The Lambda source finds all the Lambda clusters in the current AWS account with their current versions.

### New versions discovery
This source discovers new versions by reading the supported runtimes list provided by AWS in the Go Lambda SDK Library

### Default deprecated runtimes list:
AWS does not expose any API to retrieve the deprecated runtimes, therefore we use a hard-coded list which will be updated in time:
Here is the default list:
```yaml
    "nodejs",
    "nodejs4.3",
    "nodejs4.3-edge",
    "nodejs6.10",
    "nodejs8.10",
    "nodejs10.x",
    "nodejs12.x",
    "java8",
    "java8.al2",
    "python2.7",
    "python3.6",
    "dotnetcore1.0",
    "dotnetcore2.0",
    "dotnet6",
    "ruby2.5",
```

### Configuration

```yaml
sources:
    aws:
        lambda:
            enabled: true 
            request-timeout: 15s
            deprecated-runtimes-score: 100 # defaults to 100
            # deprecated-runtimes:
            #   - "nodejs"
            #   - "nodejs4.3"
            #   - "nodejs4.3-edge"
            #   - "nodejs6.10"
            #   - "nodejs8.10"
            #   - "nodejs10.x"
            #   - "nodejs12.x"
            #   - "java8"
            #   - "java8.al2"
            #   - "python2.7"
            #   - "python3.6"
            #   - "dotnetcore1.0"
            #   - "dotnetcore2.0"
            #   - "dotnet6"
            #   - "ruby2.5"
```
Parameters:
- **request-timeout**: the timeout duration when calling the AWS Lambda API
- **deprecated-runtimes-score**: arbitrarily sets the score for runtimes that are considered deprecated. It is recommended to set it to a score above your [alert threshold](../../README.md#alerting-patterns-deciding-when-to-update-softwares)  because a deprecated runtime is not maintained by AWS anymore and could stop working anytime.
- **deprecated-runtimes**: custom list of runtimes to consider deprecated (overrides the default deprecated runtimes list)

### Obsolescence Score Calculation
The Lambda source is using the [Candidate Count Calculator](../calculators/candidate_count_calculator.md)