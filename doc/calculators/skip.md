# Skip Calculator

## Description
The Skip Calculator arbitrarily gives an obsolescence score of `0`

## Example
An AWS Lambda function is using a deprecated runtime (a runtime can be evaluated as deprecated because AWS marked it as deprecated or because the is has been added to the list of deprecated runtimes in `config.sources.aws.lambda.deprecate-runtimes`).
Let's say the runtime is Python 2.7.

In that case, the lambda source will arbitrarily set the score to the value of `config.sources.aws.lambda.deprecated-runtimes-score` (defaults to `100`).