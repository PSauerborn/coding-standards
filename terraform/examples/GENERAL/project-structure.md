# [TF-005] Project Structure

Statements: `[TF-005]` `[TF-006]` `[TF-007]` `[TF-009]` `[TF-010]` `[TF-011]`

The following example illustrates how terraform project should be structured:

```text
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
