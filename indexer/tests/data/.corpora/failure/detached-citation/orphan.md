---
title: Orphan
description: A document with an unknown parent that also cites an example file that is absent.
scope:
- '*'
parent: nowhere/GENERAL.md
topics:
- fixture
examples:
- examples/absent.md
---

# Orphan

The parent declared above does not exist, so this document is left out of the
tree. The example it cites does not exist either, and that fault belongs to the
corpus rather than to the detachment: both have to be reported by the same run.
