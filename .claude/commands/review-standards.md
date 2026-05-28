---
description: Review standards for inconsistencies, and adherence to best practices.
---

You are a coding agent working on a production-grade, multi-environment platform, and you've been assigned a ticket to execute a well-scoped code change.

Review all the standards markdown files in this repository and provide feedback based on the following criteria:

1. clarity - are standards clear and easy to follow?
2. consistency - are practices and instructions consistent across different languages and stacks?
3. readability -  are standards and instructions provided in a format that is optimized for parsing/traversal by an agent?
4. length - is there enough provided context to ensure an agent can implement the logic, but not so much that it fills an agents context window with unnecessary information? 
5. conciseness - are standards written in such a way that the amount of text is minimized? can any rules be combined or removed?
6. best practices - does the format of the standards and instructions align with modern best practices? Focus on the how information is structured and presented rather than the specific rules/instructions. This does not apply to meta rules that define agent behavior.
7. mistakes - check for spelling and grammatical errors. review any code blocks for syntax errors. do not flag missing imports as many examples are partial snippets.

Ignore the following files:

- @CLAUDE.md
- @README.md
- @PLAN.md
- @REVIEW.md

Write the results of your review to @REVIEW.md.

$ARGUMENTS
