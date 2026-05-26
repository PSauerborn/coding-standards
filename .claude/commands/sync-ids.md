---
description: Regenerate and align statement IDs across all standards documents.
---

Review all standards markdown files present in this repository and make sure that 

- all statements have an ID
- all statement IDs are unique
- statement IDs are ordered correctly. 

It should be clear from the ID which language/framework the rule is referring to. For example, a golang rule could have rule ID `GO-123`. If a file already has a convention/ID format, make sure to following the format of the existing IDs rather than creating a new format.

If all of the above criteria is satisfied for all files, indicate to the user that no changes are required.

Ignore the following files:

- @CLAUDE.md
- @README.md
- @PLAN.md

$ARGUMENTS
