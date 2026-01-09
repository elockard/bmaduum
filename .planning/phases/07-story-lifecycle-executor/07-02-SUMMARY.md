---
phase: 07-story-lifecycle-executor
plan: 02
subsystem: lifecycle
tags: [executor, orchestration, workflow, tdd]

# Dependency graph
requires:
  - phase: 06-lifecycle-definition
    provides: GetLifecycle function returning workflow sequence with NextStatus
  - phase: 07-01
    provides: Writer.UpdateStatus for status persistence
provides:
  - Executor struct orchestrating full story lifecycle
  - WorkflowRunner, StatusReader, StatusWriter interfaces
  - Execute function running all workflows with status updates
affects:
  [08-update-run-command, 09-update-epic-command, 10-update-queue-command]

# Tech tracking
tech-stack:
  added: []
  patterns: [interface-based-dependency-injection, fail-fast-execution]

key-files:
  created: [internal/lifecycle/executor.go, internal/lifecycle/executor_test.go]
  modified: []

key-decisions:
  - "Interface-based DI for WorkflowRunner, StatusReader, StatusWriter"
  - "Fail-fast on workflow failure or status update failure"
  - "Returns router.ErrStoryComplete for done stories"

patterns-established:
  - "Lifecycle orchestration: status→workflow→update→repeat"
  - "Interface extraction for testability of concrete types"

issues-created: []

# Metrics
duration: 2min
completed: 2026-01-09
---

# Phase 7 Plan 02: Lifecycle Executor Summary

**Executor struct orchestrating full story lifecycle (status→workflow→update→repeat) with interface-based DI and fail-fast behavior**

## Performance

- **Duration:** 2 min
- **Started:** 2026-01-09T02:04:27Z
- **Completed:** 2026-01-09T02:06:32Z
- **Tasks:** RED → GREEN (no refactor needed)
- **Files modified:** 2

## Accomplishments

- Executor struct with WorkflowRunner, StatusReader, StatusWriter interfaces
- Execute function runs complete lifecycle from current status to done
- Fail-fast on workflow failure or status update failure
- Comprehensive test coverage with mock implementations

## Task Commits

TDD cycle commits:

1. **RED: Failing tests** - `bf9b5f8` (test)
2. **GREEN: Implementation** - `83078ec` (feat)

**Plan metadata:** (this commit)

## RED Phase

Tests written for:

- Story from backlog runs full lifecycle (4 workflows, 4 status updates)
- Story from ready-for-dev skips create-story (3 workflows)
- Story from review runs code-review and git-commit (2 workflows)
- Story already done returns router.ErrStoryComplete
- Get status error propagates
- Workflow failure stops execution (fail-fast)
- Status update failure stops execution (fail-fast)

Tests failed because Execute() returned nil without doing anything.

## GREEN Phase

Implementation:

1. Get current story status via StatusReader
2. Get lifecycle steps via router.GetLifecycle(status)
3. For each step: run workflow, update status
4. Return nil on success, error on failure

Clean, minimal implementation - no refactoring needed.

## Files Created/Modified

- `internal/lifecycle/executor.go` - Executor struct, NewExecutor, Execute, interfaces
- `internal/lifecycle/executor_test.go` - MockWorkflowRunner, MockStatusReader, MockStatusWriter, TestNewExecutor, TestExecute

## Decisions Made

- Interface-based DI: WorkflowRunner, StatusReader, StatusWriter interfaces allow easy mocking
- Fail-fast: First workflow or status update failure stops execution
- Returns router.ErrStoryComplete for done stories (consistent with router package)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- Lifecycle executor complete
- Phase 7 complete: Status Writer (07-01) + Lifecycle Executor (07-02)
- Ready for Phase 8 (Update Run Command) to integrate executor into CLI
- workflow.Runner already satisfies WorkflowRunner interface
- status.Reader already satisfies StatusReader interface
- status.Writer already satisfies StatusWriter interface

---

_Phase: 07-story-lifecycle-executor_
_Completed: 2026-01-09_
