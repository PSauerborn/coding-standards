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
- any file under an `examples/` directory — these are companion example files that contain no statements and must not receive new IDs.

Example files do reference existing statement IDs in their `# [<ID>] <Title>` heading and their `Statements:` line. Whenever a statement ID is renumbered, update every example heading and `Statements:` line that refers to it so that the references stay accurate. Example filenames are topic-based and must not change during renumbering.

$ARGUMENTS
