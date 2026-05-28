---
title: Logging General Standards
description: General standards for logging implementation.
scope:
- '*'
parent: GENERAL.md
topics:
- logging
---

# Logging General Standards

## 1. General Guidelines

`[LOG-001]` **MUST**: All applications must implement logging to provide visibility into runtime behavior and support debugging, monitoring, and incident response.

`[LOG-002]` **SHOULD**: Logging should be implemented using a structured logging package to ensure that logs can be easily stored and searched. By default, logs should be structured in JSON format.

`[LOG-003]` **SHOULD**: Structured logs should contain at minimum the following fields:

- message
- timestamp
- level

`[LOG-004]` **SHOULD**: Context-specific information should be provided in a dedicated field, not in the log message itself to increase searchability.

`[LOG-005]` **SHOULD**: Logs should be written to both standard output (stdout) and a log file. This ensures that recent logs can be easily inspected, and historic logs remain accessible.

`[LOG-006]` **SHOULD**: Log levels should be configurable via environment variables to ensure that logging verbosity can be controlled without making code changes.

### Example

The following example illustrates how logs should be structured

```txt
<!-- GOOD -->
{"message": "received request to get user", "timestamp": 1779914598, "level": "info", "username": "TEST_USER_1"}
{"message": "received request to update user", "timestamp": 1779916598, "level": "info", "username": "TEST_USER_1", "updated_field": "first_name"}

```

The following example illustrates how logs should NOT be structured

```txt
<!-- BAD -->
<!-- BAD: context-specific information stored in message rather than dedicated field -->
{"message": "received request to get user TEST_USER", "timestamp": 1779914598, "level": "info"}
<!-- BAD: missing required fields -->
{"message": "received request to update user", "level": "info", "username": "TEST_USER_1", "updated_field": "first_name"}

```