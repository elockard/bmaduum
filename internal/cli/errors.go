package cli

import "fmt"

// ExitError represents a command execution failure with a specific exit code.
//
// This error type allows Cobra RunE functions to signal non-zero exit codes
// without calling os.Exit() directly, enabling testable CLI behavior.
// When a command fails, it returns NewExitError(code), which propagates up
// to [RunWithConfig] where [IsExitError] extracts the code for [ExecuteResult].
//
// Testability benefit: Tests can assert on exit codes without process termination.
// The [Execute] function handles the actual os.Exit() call based on the code.
type ExitError struct {
	// Code is the exit code to return to the shell.
	// Convention: 0 = success, 1 = general error, other values from subprocess.
	Code int
}

// Error implements the error interface, returning a string in the format
// "exit status N" where N is the exit code. This format matches the standard
// os/exec ExitError format for consistency with subprocess exit messages.
func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

// NewExitError creates an [ExitError] with the given exit code.
//
// Use this in Cobra RunE functions to signal failure:
//
//	if err != nil {
//	    return NewExitError(1)  // or pass through subprocess exit code
//	}
//
// The code parameter is typically 1 for CLI errors, or passed through from
// a failed subprocess (e.g., Claude CLI returning non-zero).
func NewExitError(code int) *ExitError {
	return &ExitError{Code: code}
}

// IsExitError checks if an error is an [ExitError] and extracts its exit code.
//
// Returns (code, true) if err is an *ExitError, allowing the caller to handle
// the specific exit code. Returns (0, false) for nil or non-ExitError errors.
//
// Typical usage in [RunWithConfig]:
//
//	if err := cmd.Execute(); err != nil {
//	    if code, ok := IsExitError(err); ok {
//	        return ExecuteResult{ExitCode: code, Err: err}
//	    }
//	    return ExecuteResult{ExitCode: 1, Err: err}  // generic error
//	}
func IsExitError(err error) (int, bool) {
	if exitErr, ok := err.(*ExitError); ok {
		return exitErr.Code, true
	}
	return 0, false
}
