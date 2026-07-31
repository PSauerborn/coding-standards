# [LOG-002] Structured Log Format

Statements: `[LOG-002]` `[LOG-003]` `[LOG-004]`

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
