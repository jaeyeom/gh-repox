package cli

import "fmt"

// Exit codes as documented in README.
const (
	ExitRuntime      = 1 // Runtime error (execution failed)
	ExitInvalidInput = 2 // Invalid input or configuration
	ExitNoAuth       = 3 // GitHub authentication unavailable
	ExitCreateFailed = 4 // Repository creation failed
	ExitStrictFailed = 5 // Post-create or apply settings failed in strict mode
	ExitCloneFailed  = 6 // Clone failed in strict mode
)

// ExitError is an error that carries a specific exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

func exitErrorf(code int, format string, args ...any) *ExitError {
	return &ExitError{Code: code, Err: fmt.Errorf(format, args...)}
}
