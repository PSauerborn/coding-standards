# CLAUDE.md

Global `CLAUDE.md` file containing instructions on how agents should process tasks. The instructions provided in this file __MUST__ be followed unless explicitly overridden by either a local `CLAUDE.md` file, or a users instruction.

## Coding Standards

When working on a task, consult the standards tree in `/Users/Pascal/.stdidx/standards-tree.yaml`
to find applicable coding standards.

1. Always start at the root nodes. Read any root node whose scope
   matches the files you're working with or whose scope is "*".

2. For each node you read, check its children. Descend into a child
   if its scope or tags match your current context.

3. Stop descending a branch when no children match your context.

4. Collect all matching nodes from root to leaf. Standards at every
   level in the path apply — a child does not replace its parent,
   it adds to it.

5. If a child standard contradicts a parent, the child takes precedence.

To determine if a node matches your context:
- description: compare the description of the node to the task you're working on
- scope: compare against the file extensions you're editing
  ("*.py", "*.ts", "*" matches everything)
- topics: compare against the project's detected frameworks/tools
  (e.g. if package.json has "react" as a dependency, the "react"
  topic matches)

## Development

Review coding standards before making any code changes. 

Development should be done in separate planning and implementation phases. For any task involving more than one file or more than ~20 lines of code, always write an implementation plan to `PLAN.md` first and explicitly wait for approval before proceeding. Once approved, proceed with implementation.

Implementation plans should contain the following at minimum:

- Files to modify
- Steps to take
- Expected changes

The goal is to create an implementation plan with enough detail that an agent with a smaller thinking capacity can execute the changes effectively.

Updates to existing code should be accompanied with a markdown `diff` block that summarizes the changes being made. This ensures that the changes are easily reviewable. Do not generate the diff block for new files.

Tests should be ran after each change to ensure that the changes are correct.
