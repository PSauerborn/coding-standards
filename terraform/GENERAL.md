---
title: Terraform Code Standards
description: Standards for Terraform project structure and state management.
scope:
- '*.tf'
parent: GENERAL.md
topics:
- terraform
- iac
examples:
- examples/GENERAL/project-structure.md
- examples/GENERAL/provider-state.md
---

# Terraform Code Standards

## 1. Versions and Tooling

`[TF-001]` **MUST**: Use terraform version `1.11.4` or higher.

`[TF-002]` **MUST**: Terraform must be formatted with `terraform fmt`

`[TF-003]` **MUST**: Terraform must be validated with `terraform validate`

`[TF-004]` **MUST**: Terraform must be scanned for security vulnerabilities using `checkov` or `trivy`.

## 2. Syntax, Naming & Style

`[TF-005]` **MUST**: Terraform must be implemented in line with the Google best practices. At minimum, this includes a `modules/main` directory and a `env/{env}` directory for each environment. By default, there should be a `dev` and `prod` environment. See `examples/GENERAL/project-structure.md` for implementation.

`[TF-006]` **MUST**: Each environment folder must have at minimum a `provider.tf` file that defines the required providers, a `main.tf` file that invokes the main module found at `modules/main`, and a `outputs.tf` file that defines all outputs used by the environment. See `examples/GENERAL/project-structure.md` for implementation.

`[TF-007]` **MUST**: The `modules/main` directory must have at minimum a `main.tf` file that defines the main module, a `versions.tf` file that defines the required Terraform version and the required providers, and a `variables.tf` file that defines all variables used by the module.

`[TF-008]` **MUST**: The `modules/main/variables.tf` file must contain an `environment` variable that is used to define the environment the module is being deployed to.

`[TF-009]` **MUST**: Each module must have a `versions.tf` file that defines the required Terraform version and the required providers.

`[TF-010]` **MUST**: Each module must have a `variables.tf` file that defines all variables used by the module.

`[TF-011]` **MUST**: Each module must have a `outputs.tf` file that defines all outputs used by the module.

`[TF-012]` **SHOULD**: Each module should have a `main.tf`. This should contain a `locals` block that defines a `base_name` that can be used as a prefix for all resources in the module. This should contain the value of the `environment` variable to make sure that resources are not overwritten in different environments.

`[TF-013]` **SHOULD**: Avoid creating dependencies between terraform modules that are not in the same state file. Self-contained modules should always be preferred as it reduces the number of dependencies between terraform components.

## 3. Configuration

`[TF-014]` **MUST**: Terraform state must be stored in an S3 bucket. See `examples/GENERAL/provider-state.md` for implementation.

`[TF-015]` **MUST**: Terraform statelocks must be implemented using an S3 lockfile. See `examples/GENERAL/provider-state.md` for implementation.

`[TF-016]` **SHOULD**: A dedicated role should be assumed when accessing AWS resources for provider configuration. See `examples/GENERAL/provider-state.md` for implementation.
