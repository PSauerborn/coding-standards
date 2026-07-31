---
title: Alpha
description: A document whose parent is Beta, which in turn declares Alpha as its parent.
scope:
- '*'
parent: BETA.md
topics:
- fixture
---

# Alpha

This document and `BETA.md` declare each other as their parent, so neither can
be placed under a root, which is the failure this fixture exists to report.
