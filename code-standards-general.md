# General Code Standards

This document contains coding standards for general code. The document outlines a series of **MUST** and **SHOULD**. **MUST**s** are mandatory and must be followed. **SHOULD** are best practices and should be implemented where reasonable. Examples should be treated as **SHOULD**.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

## General

**SHOULD**: All design and implementation choices should follow KISS (Keep It Simple, Stupid) and YAGNI (You Ain't Gonna Need It) principles.

**SHOULD**: Complexity intruduces significant cost, engineering debt and risk. Prefer solutions and implementations that minimize entropy.

**SHOULD**: Code should be built, tested and ran in a containerized environment, preferably using Docker. This ensures consistency and reproducibility.

**SHOULD**: Makefiles should be used to define build and test targets.

**SHOULD**: `pre-commit` hooks should be used to enforce coding standards and best practices.

**SHOULD**: Projects should have decidcated integration tests.
