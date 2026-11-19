# [GO-004] golangci-lint Configuration

Statements: `[GO-004]` `[GO-046]` `[GO-047]`

The following `.golangci.yml` sits at the module root of every Go application.

```yaml
# GOOD
# File: .golangci.yml
version: "2"

linters:
  default: standard
  enable:
    # GO-047: blank-line discipline. `golangci-lint run --fix` applies it.
    - wsl_v5
    # GO-046: a production function scoring 15 or more fails the lint.
    - gocognit
  settings:
    gocognit:
      # gocognit reports a function whose complexity is above this value.
      min-complexity: 14
  exclusions:
    rules:
      # Table-driven tests nest t.Run closures and score high by design.
      - path: _test\.go
        linters:
          - gocognit
```
