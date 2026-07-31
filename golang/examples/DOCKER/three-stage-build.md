# [GO-DOCKER-001] Three-Stage Distroless Build

Statements: `[GO-DOCKER-001]` `[GO-DOCKER-002]`

The following dockerfile illustrates a three-stage build for a Go application using the distroless runtime image.

```dockerfile
# GOOD: Use golang:1.25 as base image for tests
# GOOD: implement unittests as first stage
FROM golang:1.25 AS tests

WORKDIR /app/tests

COPY go.mod go.sum ./
RUN go mod download

RUN go install gotest.tools/gotestsum@latest

COPY etc ./etc
COPY *.go ./

CMD ["gotestsum", "--format", "testname"]

# GOOD: Use golang:1.25 as base image for build
# GOOD: implement build as second stage
FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o api .

# GOOD: Use gcr.io/distroless/static:nonroot as base image for runtime
# GOOD: fall back to debian:*-slim only if the binary cannot run on distroless
FROM gcr.io/distroless/static:nonroot AS runtime

WORKDIR /app

COPY --from=build /app/api .

COPY etc ./etc

CMD ["./api"]
```
