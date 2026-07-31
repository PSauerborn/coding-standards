---
title: Beta
description: A document whose parent is Alpha, which in turn declares Beta as its parent.
scope:
- '*'
parent: ALPHA.md
topics:
- fixture
---

# Beta

This document and `ALPHA.md` declare each other as their parent, so neither can
be placed under a root, which is the failure this fixture exists to report.
