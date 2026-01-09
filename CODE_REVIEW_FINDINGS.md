# Comprehensive Code Review Findings

**Project:** bmad-automate
**Review Date:** 2026-01-08
**Codebase:** 6,418 LOC Go
**Go Version:** 1.25.5

---

## Executive Summary

| Dimension         | Score            | Rating             |
| ----------------- | ---------------- | ------------------ |
| Code Quality      | 8.2/10           | Good               |
| Architecture      | B+               | Good               |
| Security          | Medium-High Risk | Requires Attention |
| Performance       | Good             | Acceptable         |
| Test Coverage     | 73.4%            | Good               |
| Documentation     | A- (91/100)      | Excellent          |
| Go Best Practices | 92/100           | Excellent          |
| CI/CD & DevOps    | 2.1/5            | Needs Improvement  |

---

## Critical Issues (P0)

### SEC-01: Dangerous Permission Bypass by Design

- **Severity:** CRITICAL (CVSS 9.8)
- **Location:** `internal/claude/client.go:77-78, 114-115`
- **Issue:** Tool unconditionally passes `--dangerously-skip-permissions` to Claude CLI, bypassing all permission checks
- **Impact:** If Claude CLI is compromised or prompts are crafted maliciously, attacker could gain arbitrary code execution on the host system
- **Current Code:**
  ```go
  cmd := exec.CommandContext(ctx, e.config.BinaryPath,
      "--dangerously-skip-permissions",
      "-p", prompt,
      "--output-format", e.config.OutputFormat,
  )
  ```
- **Remediation:**
  - [ ] Add prominent security warnings to README.md
  - [ ] Add warnings to CLI help text
  - [ ] Consider adding `--force` flag to make dangerous mode opt-in
  - [ ] Implement operation allowlist for Claude actions
  - [ ] Add audit logging of all operations performed

---

### SEC-02: Arbitrary Command Execution via `raw` Command

- **Severity:** HIGH (CVSS 8.6)
- **Location:** `internal/cli/raw.go:19-22`
- **Issue:** The `raw` command accepts any arbitrary text and passes it directly to Claude CLI with dangerous permissions
- **Current Code:**
  ```go
  prompt := strings.Join(args, " ")
  exitCode := app.Runner.RunRaw(ctx, prompt)
  ```
- **Remediation:**
  - [ ] Restrict `raw` command to development/debug builds only
  - [ ] Add authentication/authorization before allowing raw prompts
  - [ ] Add rate limiting and audit logging

---

## High Priority (P1)

### SEC-06: Binary Path Injection

- **Severity:** HIGH (CVSS 7.8)
- **Location:** `internal/config/config.go:68-70`
- **Issue:** Claude binary path can be overridden via `BMAD_CLAUDE_PATH` environment variable
- **Current Code:**
  ```go
  if binaryPath := os.Getenv("BMAD_CLAUDE_PATH"); binaryPath != "" {
      cfg.Claude.BinaryPath = binaryPath
  }
  ```
- **Remediation:**
  - [ ] Validate binary path exists and is executable
  - [ ] Check binary signature/hash against known good values
  - [ ] Restrict path to absolute paths within known directories
  - [ ] Log when non-default binary paths are used

---

### SEC-05: Configuration File Hijacking

- **Severity:** MEDIUM (CVSS 5.5)
- **Location:** `internal/config/config.go:43-50`
- **Issue:** Configuration loaded from current directory could be attacker-controlled
- **Remediation:**
  - [ ] Restrict config file permissions (0600)
  - [ ] Validate config file ownership matches current user
  - [ ] Use absolute paths for production deployments
  - [ ] Log when custom config paths are used

---

### CICD-01: No CI/CD Pipeline Exists

- **Severity:** HIGH
- **Location:** `.github/workflows/` (missing)
- **Issue:** Project has no automated testing, linting, or release pipeline
- **Remediation:**
  - [ ] Create `.github/workflows/ci.yml` with:
    - Build verification
    - Test execution with race detection
    - golangci-lint
    - govulncheck security scanning
  - [ ] Add pre-commit hooks
  - [ ] Add goreleaser for automated releases

**Recommended CI workflow:**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Verify dependencies
        run: go mod verify
      - name: Build
        run: go build -v ./...
      - name: Test
        run: go test -v -race -coverprofile=coverage.out ./...
      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Vulnerability check
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
```

---

### CODE-01: Duplicate Dry-Run Code

- **Severity:** MEDIUM
- **Location:**
  - `internal/cli/queue.go:74-112`
  - `internal/cli/epic.go:84-122`
- **Issue:** Nearly identical 38-line functions for dry-run logic
- **Remediation:**
  - [ ] Extract shared function:
    ```go
    // internal/cli/dryrun.go
    func printDryRunPlan(executor *lifecycle.Executor, header string, storyKeys []string) error
    ```

---

### CODE-02: Interface Duplication

- **Severity:** MEDIUM
- **Locations:**
  - `internal/cli/root.go:24-32` (StatusReader, StatusWriter, WorkflowRunner)
  - `internal/lifecycle/executor.go:12-24` (same interfaces)
  - `internal/workflow/queue.go:15-17` (StatusReader)
- **Issue:** Same interfaces defined in 3+ places
- **Remediation:**
  - [ ] Consolidate interfaces into single location
  - [ ] Option 1: Create `internal/interfaces.go`
  - [ ] Option 2: Use interface embedding from one source

---

### TEST-01: RunQueueWithStatus Has 0% Coverage

- **Severity:** MEDIUM
- **Location:** `internal/workflow/queue.go:31`
- **Issue:** Critical function has no direct unit tests (only tested via CLI integration)
- **Remediation:**
  - [ ] Add direct unit tests for `QueueRunner.RunQueueWithStatus`
  - [ ] Test error handling when status read fails
  - [ ] Test router error handling (ErrUnknownStatus)
  - [ ] Test stop-on-failure behavior

---

### ARCH-01: Printer Interface Too Large (ISP Violation)

- **Severity:** MEDIUM
- **Location:** `internal/output/printer.go:28-58`
- **Issue:** Interface has 16 methods, violates Interface Segregation Principle
- **Remediation:**
  - [ ] Split into focused sub-interfaces:

    ```go
    type SessionPrinter interface {
        SessionStart()
        SessionEnd(duration time.Duration, success bool)
    }

    type StepPrinter interface {
        StepStart(step, total int, name string)
        StepEnd(duration time.Duration, success bool)
    }

    type ToolPrinter interface {
        ToolUse(name, description, command, filePath string)
        ToolResult(stdout, stderr string, truncateLines int)
    }

    type Printer interface {
        SessionPrinter
        StepPrinter
        ToolPrinter
        // ... other methods
    }
    ```

---

## Medium Priority (P2)

### CODE-03: Redundant Success Field

- **Location:** `internal/workflow/steps.go:13-19`
- **Issue:** `StepResult.Success` field is redundant with `IsSuccess()` method
- **Current Code:**

  ```go
  type StepResult struct {
      Name     string
      Duration time.Duration
      ExitCode int
      Success  bool  // Redundant
  }

  func (r StepResult) IsSuccess() bool {
      return r.ExitCode == 0
  }
  ```

- **Remediation:**
  - [ ] Remove `Success` field and use `IsSuccess()` method only, OR
  - [ ] Remove `IsSuccess()` method and use field directly

---

### CODE-04: Duplicate StepResult Type

- **Locations:**
  - `internal/workflow/steps.go:13`
  - `internal/output/printer.go:12`
- **Issue:** Two slightly different `StepResult` types exist
- **Remediation:**
  - [ ] Create shared type in a common package
  - [ ] Option: Create `internal/domain/types.go`

---

### PERF-01: Template Re-parsing on Every Call

- **Location:** `internal/config/config.go:109-121`
- **Issue:** Templates are parsed every time `GetPrompt()` is called
- **Current Code:**
  ```go
  func expandTemplate(tmpl string, data PromptData) (string, error) {
      t, err := template.New("prompt").Parse(tmpl)  // Parsed on every call
      // ...
  }
  ```
- **Remediation:**
  - [ ] Parse templates during `Config` initialization
  - [ ] Store parsed templates in map
  - [ ] Cache in `Config` struct

---

### PERF-02: Unbuffered Terminal Output

- **Location:** `internal/output/printer.go:66-67, 75-77`
- **Issue:** Every `writeln` call triggers unbuffered write to stdout
- **Remediation:**
  - [ ] Wrap `os.Stdout` with `bufio.Writer`
  - [ ] Flush at session boundaries

---

### PERF-03: Double JSON Allocation

- **Location:** `internal/claude/parser.go:48-54`
- **Issue:** Creates redundant string copies during JSON parsing
- **Current Code:**
  ```go
  line := scanner.Text()        // Copy #1: []byte -> string
  // ...
  json.Unmarshal([]byte(line), &streamEvent)  // Copy #2: string -> []byte
  ```
- **Remediation:**
  - [ ] Use `scanner.Bytes()` directly, OR
  - [ ] Use `json.NewDecoder` for streaming

---

### PERF-04: Parser Goroutine Lacks Context Cancellation

- **Location:** `internal/claude/parser.go:31-68`
- **Issue:** Parser goroutine has no cancellation mechanism
- **Remediation:**
  - [ ] Accept `context.Context` parameter
  - [ ] Select on `ctx.Done()` in send loop:
    ```go
    select {
    case <-ctx.Done():
        return
    case events <- event:
    }
    ```

---

### GO-01: Type Assertions Should Use errors.As

- **Locations:**
  - `internal/cli/errors.go:24-29`
  - `internal/claude/client.go:150`
  - `internal/config/config.go:54-56`
- **Issue:** Direct type assertions don't work with wrapped errors
- **Current Code:**
  ```go
  if exitErr, ok := err.(*ExitError); ok {
  ```
- **Remediation:**
  - [ ] Use `errors.As` for wrapped error compatibility:
    ```go
    var exitErr *ExitError
    if errors.As(err, &exitErr) {
        return exitErr.Code, true
    }
    ```

---

### SEC-03: Template Injection via StoryKey

- **Severity:** MEDIUM (CVSS 6.5)
- **Location:** `internal/config/config.go:109-120`
- **Issue:** Story keys inserted into Go templates without validation
- **Remediation:**
  - [ ] Add regex validation for story key format
  - [ ] Example: `^[A-Z0-9]+-[0-9]+(-[a-z0-9-]+)?$`
  - [ ] Sanitize `{{` and `}}` sequences

---

### DOC-01: PACKAGES.md Missing New Packages

- **Location:** `docs/PACKAGES.md`
- **Issue:** Missing documentation for v1.1 packages
- **Remediation:**
  - [ ] Add `internal/lifecycle` package documentation
  - [ ] Add `internal/state` package documentation

---

### DOC-02: ARCHITECTURE.md Needs Update

- **Location:** `docs/ARCHITECTURE.md`
- **Issue:** Doesn't reflect v1.1 architecture changes
- **Remediation:**
  - [ ] Add lifecycle package to layer diagram
  - [ ] Update package dependency diagram
  - [ ] Document state management

---

### DOC-03: Missing TROUBLESHOOTING.md

- **Location:** `docs/` (missing)
- **Remediation:**
  - [ ] Create TROUBLESHOOTING.md with:
    - Common errors and solutions
    - Claude CLI integration issues
    - Configuration problems

---

### DOC-04: Missing CHANGELOG.md

- **Location:** Root directory (missing)
- **Remediation:**
  - [ ] Create CHANGELOG.md
  - [ ] Document v1.0 release
  - [ ] Document v1.1 changes

---

## Low Priority (P3)

### CODE-05: Exported Constants Only Used Internally

- **Locations:**
  - `internal/state/state.go:12` - `StateFileName`
  - `internal/status/reader.go:15` - `DefaultStatusPath`
- **Remediation:**
  - [ ] Make unexported: `stateFileName`, `defaultStatusPath`

---

### CODE-06: Missing String() Method on LifecycleStep

- **Location:** `internal/router/lifecycle.go`
- **Issue:** Step display format duplicated in multiple places
- **Remediation:**
  - [ ] Add method:
    ```go
    func (s LifecycleStep) String() string {
        return fmt.Sprintf("%s -> %s", s.Workflow, s.NextStatus)
    }
    ```

---

### PERF-05: UTF-8 Truncation May Corrupt Characters

- **Location:** `internal/output/printer.go:277-282`
- **Issue:** Byte-based truncation may split multi-byte UTF-8 characters
- **Current Code:**
  ```go
  return s[:maxLen-3] + "..."  // May corrupt UTF-8
  ```
- **Remediation:**
  - [ ] Use `utf8.RuneCountInString` and proper rune slicing

---

### TEST-02: Unused Variable in Test

- **Location:** `internal/workflow/workflow_test.go:98-99`
- **Issue:** Unused variable assignment
- **Current Code:**
  ```go
  originalExecute := mockExecutor.ExecuteWithResult
  _ = originalExecute
  ```
- **Remediation:**
  - [ ] Remove unused variable

---

### CICD-02: justfile Missing lint in check

- **Location:** `justfile`
- **Issue:** `check` command runs fmt/vet/test but not lint
- **Remediation:**
  - [ ] Update to: `check: fmt vet lint test`

---

### CICD-03: Missing Pre-commit Hooks

- **Location:** `.git/hooks/` (only .sample files)
- **Remediation:**
  - [ ] Create `.git/hooks/pre-commit`:
    ```bash
    #!/bin/sh
    just check
    ```

---

### CICD-04: Missing govulncheck

- **Location:** `justfile`
- **Remediation:**
  - [ ] Add task:
    ```
    vulncheck:
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...
    ```

---

### CICD-05: No Cross-Platform Builds

- **Location:** `justfile`
- **Remediation:**
  - [ ] Add build-all task for darwin/linux/windows on amd64/arm64

---

### CICD-06: No Version Injection

- **Location:** `justfile`
- **Remediation:**
  - [ ] Add ldflags for version:
    ```
    build:
        go build -ldflags "-X main.version=$(git describe --tags) -X main.commit=$(git rev-parse --short HEAD)" -o {{binary_name}} ./cmd/bmad-automate
    ```

---

## Test Coverage Summary

| Package            | Coverage  | Status     | Notes                        |
| ------------------ | --------- | ---------- | ---------------------------- |
| internal/lifecycle | 100.0%    | Excellent  | -                            |
| internal/router    | 100.0%    | Excellent  | -                            |
| internal/output    | 96.4%     | Excellent  | -                            |
| internal/config    | 89.1%     | Good       | -                            |
| internal/status    | 89.6%     | Good       | -                            |
| internal/state     | 83.3%     | Good       | -                            |
| internal/claude    | 62.5%     | Needs Work | Add DefaultExecutor tests    |
| internal/cli       | 61.5%     | Needs Work | -                            |
| internal/workflow  | 51.1%     | Needs Work | Add RunQueueWithStatus tests |
| cmd/bmad-automate  | 0.0%      | Expected   | Thin wrapper                 |
| **Total**          | **73.4%** | **Good**   | Target: 80%                  |

---

## Security Risk Matrix

| ID     | Vulnerability         | Severity | CVSS | Status    | Action   |
| ------ | --------------------- | -------- | ---- | --------- | -------- |
| SEC-01 | Permission Bypass     | CRITICAL | 9.8  | By Design | Document |
| SEC-02 | Arbitrary Execution   | HIGH     | 8.6  | Open      | Review   |
| SEC-06 | Binary Path Injection | HIGH     | 7.8  | Open      | Fix      |
| SEC-05 | Config File Hijacking | MEDIUM   | 5.5  | Open      | Fix      |
| SEC-03 | Template Injection    | MEDIUM   | 6.5  | Mitigated | Validate |
| SEC-04 | Path Traversal        | LOW      | 4.3  | Mitigated | -        |

---

## Implementation Checklist

### Phase 1: Immediate (This Week)

- [ ] SEC-01: Add security warnings to README.md
- [ ] CICD-01: Create `.github/workflows/ci.yml`
- [ ] CICD-03: Add pre-commit hooks

### Phase 2: Short-term (This Month)

- [ ] SEC-06: Validate binary path
- [ ] SEC-05: Add config file security checks
- [ ] CICD-04: Add govulncheck
- [ ] DOC-01: Update PACKAGES.md
- [ ] DOC-02: Update ARCHITECTURE.md
- [ ] TEST-01: Add RunQueueWithStatus tests

### Phase 3: Medium-term (This Quarter)

- [ ] ARCH-01: Split Printer interface
- [ ] CODE-01: Extract shared dry-run function
- [ ] CODE-02: Consolidate interface definitions
- [ ] DOC-03: Create TROUBLESHOOTING.md
- [ ] DOC-04: Create CHANGELOG.md
- [ ] PERF-01: Cache parsed templates
- [ ] GO-01: Use errors.As in 3 locations

### Phase 4: Low Priority (Backlog)

- [ ] CODE-03: Remove redundant Success field
- [ ] CODE-04: Consolidate StepResult types
- [ ] CODE-05: Make constants unexported
- [ ] CODE-06: Add LifecycleStep.String()
- [ ] PERF-02: Buffer terminal output
- [ ] PERF-03: Eliminate double JSON allocation
- [ ] PERF-04: Add context cancellation to parser
- [ ] PERF-05: Fix UTF-8 truncation
- [ ] TEST-02: Remove unused test variable
- [ ] CICD-02: Add lint to check command
- [ ] CICD-05: Add cross-platform builds
- [ ] CICD-06: Add version injection

---

## Notes

- The `--dangerously-skip-permissions` flag is an intentional design decision for automation. The security findings are about documentation and hardening, not removing this functionality.
- Test coverage of 73.4% is good for a CLI tool; target 80% by adding tests for workflow/queue.go and claude/client.go.
- The lack of CI/CD is the most significant operational gap and should be addressed first.

---

_Generated: 2026-01-08_
