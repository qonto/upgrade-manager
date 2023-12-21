# Sources

A `Source` represent a type of software that can be automatically discovered by `upgrade-manager`. 

The `Source` is responsible for two tasks:
1. Automatically `discover instances of a given software type`. For example it could find all ArgoCD applications using helm in a kubernetes cluster, or all AWS Lambda functions deployed in an AWS Account.
2. Automatically `discover version candidates for the software`, meaning newer versions that are eligible as per the source's filtering logic (see more details in each source's documentation),

Here is the existing list of currently supported sources:
 - [ArgoCD Helm applications (ArgoCDHelm)](./argoCDHelm.md)
 - [Local Helm directory (filesystemhelm)](./filesystemhelm.md)
 - [AWS Lambda](./aws_lambda.md)
 - [AWS MSK](./aws_msk.md)
 - [AWS Elasticache](./aws_elasticache.md)
 - [AWS EKS](./aws_eks.md)
 - [AWS RDS](./aws_rds.md)
