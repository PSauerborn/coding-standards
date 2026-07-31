# [TF-014] Provider and State Configuration

Statements: `[TF-014]` `[TF-015]` `[TF-016]`

The following example illustrates how terraform providers should be structured with S3 state storage and DynamoDB statelocks.

```terraform
# GOOD
# File: env/dev/provider.tf

terraform {
  required_version = ">= 1.11.4"

  required_providers {}

  backend "s3" {
    key            = "state/app=example_app/env=dev/state.tfstate" # GOOD: State is stored in an S3 bucket
    bucket         = "example-bucket"
    use_lockfile   = true # GOOD: State is locked using a lockfile

    assume_role {
      role_arn = "arn:aws:iam::123456789012:role/terraform" # GOOD: A dedicated role is used to assume when accessing AWS resources
    }
  }
}
```
