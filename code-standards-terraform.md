# Terraform Code Standards

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

# 2. Syntax, Naming & Style

**MUST**: Terraform must be implemented in line with the Google best practices. At minimum, this includes a `modules/main` directory and a `env/{env}` directory for each environment. By default, there should be a `dev` and `prod` environment. See Example 1 for implementation.

**MUST**: Each environment folder must have at minimum a `provider.tf` file that defines the required providers, a `main.tf` file that invokes the main module found at `modules/main`, and a `outputs.tf` file that defines all outputs used by the environment. See Example 1 for implementation.

**MUST**: The `modules/main` directory must have at minimum a `main.tf` file that defines the main module, a `versions.tf` file that defines the required Terraform version and the required providers, and a `variables.tf` file that defines all variables used by the module.

**MUST**: The `modules/main/variables.tf` file must contain an `environment` variable that is used to define the environment the module is being deployed to.

**MUST**: Each module must have a `versions.tf` file that defines the required Terraform version and the required providers.

**MUST**: Each module must have a `variables.tf` file that defines all variables used by the module.

**MUST**: Each module must have a `outputs.tf` file that defines all outputs used by the module.

**MUST**: Terraform state must be stored in an S3 bucket. See Example 2 for implementation.

**MUST**: Terraform statelocks must be implemented using a DynamoDB table. See Example 2 for implementation.

**SHOULD**: A dedicated role should be assumed when accessing AWS resources for provider configuration. See Example 2 for implementation.

**SHOULD**: Each module should have a `main.tf`. This should contain a `locals` block that defines a `base_name` that can be used as a prefix for all resources in the module. This should contain the value of the `environment` variable to make sure that resources are not overwritten in different environments.

## Example 1

The following example illustrates how terraform project should be structured:

```
.
├── env
│   ├── dev
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   └── provider.tf
│   └── prod
│       ├── main.tf
│       ├── outputs.tf
│       └── provider.tf
├── modules
│   └── main
│       ├── main.tf
│       ├── outputs.tf
│       ├── variables.tf
│       └── versions.tf
├── .gitignore
├── README.md
```

## Example 2

The following example illustrates how terraform providers should be structured with S3 state storage and DynamoDB statelocks.

```terraform
# GOOD
# File: env/dev/provider.tf

terraform {
  required_version = ">= 1.7.1"

  required_providers {}

  backend "s3" {
    key            = "state/app=example_app/env=dev/state.tfstate" # GOOD: State is stored in an S3 bucket
    bucket         = "example-bucket"
    region         = "eu-central-1"
    dynamodb_table = "example-table" # GOOD: State is locked using a DynamoDB table

    assume_role {
      role_arn = "arn:aws:iam::123456789012:role/terraform" # GOOD: A dedicated role is used to assume when accessing AWS resources
    }
  }
}
```