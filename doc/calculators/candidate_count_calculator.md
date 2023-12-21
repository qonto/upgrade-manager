# Candidate Count Calculator

## Description
For each version candidate the software could be upgraded to, the software will receive an extra `30 obsolescence points`.

Sources using this Calculator:
- [lambda](../sources/aws_lambda.md)


### Example
An AWS lambda function is running Python `3.10`.
Let's say the latest available version of Python on AWS Lambda is Python `3.12`
In that case, the lambda software will be given an obsolescence of `30*2=60 points` (because there are version 2 candidates, namely Python 3.11 and Python 3.12)