# Calculators

Calculators represent different strategies to compute an obsolescence score for a software. 
When a [Source](../sources/README.md) finds a software, it is responsible for attaching `Calculator` to it via the `software.Calculator` struct field.

Here is the existing list of supported calculators:
 - [Semver Calculator / Augmented Semver Calculator](./semver_calculator.md)
 - [Date Calculator](./date_calculator.md)
 - [Candidate Count Calculator](./candidate_count_calculator.md)
 - [Skip](./skip.md)
