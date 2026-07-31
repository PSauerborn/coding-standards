# [PY-DOCKER-001] Two-Stage Slim Build

Statements: `[PY-DOCKER-001]` `[PY-DOCKER-002]`

The following example shows a minimal dockerfile for a python application that implements a `tests` and a `runtime` stage. Note that a smaller runtime image is used.

```dockerfile
# GOOD: Use bookworm as base image for tests
FROM python:3.13-bookworm AS tests

COPY requirements.txt ./
COPY tests/requirements.txt ./requirements-tests.txt

RUN pip install -U pip && \
    pip install -r requirements.txt && \
    pip install -r requirements-tests.txt

COPY src ./
COPY tests ./tests

CMD ["pytest", "-vv"]

# GOOD: Use slim as base image for runtime
FROM python:3.13-slim AS runtime

COPY requirements.txt ./

RUN pip install -U pip && \
    pip install -r requirements.txt

COPY src ./

CMD ["python", "src/main.py"]
```
