# [GO-029] Unittest Structure

Statements: `[GO-026]` `[GO-028]` `[GO-029]` `[GO-032]`

The following illustrates unittests for a `SomeFunction` defined in `main.go`.

```go
// GOOD
// File: main_test.go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSomeFunction(t *testing.T) {
    // GOOD: unittests should be grouped by test case and ran in their own t.Run block
    t.Run("test case 1", func(t *testing.T) {
        result, err := SomeFunction()
        expected := "expected result"

        // GOOD: unittests should make use of the assert package
        assert.NoError(t, err)
        assert.Equal(t, expected, result)
    })

    t.Run("test case 2", func(t *testing.T) {
        result, err := SomeFunction()
        assert.Error(t, err)
    })
}
```

The following testing structure should be avoided:

```go
// BAD
// File: main_test.go
import (
    "testing"
)

// BAD: unittests should be grouped by test case and not be spread across multiple functions
func TestSomeFunctionPath1(t *testing.T) {
    result, err := SomeFunction()
    if err != nil {
        t.Fatal(err)
    }

    expected := "expected result"
    // BAD: unittests should make use of the assert package
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}

// BAD: unittests should be grouped by test case and not be spread across multiple functions
func TestSomeFunctionPath2(t *testing.T) {
    result, err := SomeFunction()
    if err != nil {
        t.Fatal(err)
    }

    expected := "expected result"
    // BAD: unittests should make use of the assert package
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}

```
