---
title: Unterminated Standard
description: A document that opens a frontmatter block and never closes it.
scope: '*'
topics:
- unterminated

# Unterminated Standard

The opening delimiter sits at byte offset 0, but no closing `---` line ever
follows, so the frontmatter block cannot be delimited.
