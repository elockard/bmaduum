---
phase: 09-update-epic-command
plan: 01
subsystem: cli
tags: [go, cobra, lifecycle, epic, tdd]

# Dependency graph
requires:
  - phase: 07-story-lifecycle-executor
    provides: lifecycle.Executor for orchestrating full story workflows
  - phase: 08-update-run-command
    provides: Pattern for lifecycle executor usage in CLI commands
provides:
  - Epic command runs full lifecycle per story
  - Interface-based DI for testability
  - Table-driven TDD tests for lifecycle behavior
affects: [queue-command, error-recovery]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - lifecycle.Executor usage in epic command
    - Table-driven lifecycle tests with MockWorkflowRunner/MockStatusWriter

key-files:
  created: []
  modified:
    - internal/cli/epic.go
    - internal/cli/epic_test.go

key-decisions:
  - "Removed legacy tests - obsolete after lifecycle change"

patterns-established:
  - "Epic command lifecycle execution follows run command pattern"

issues-created: []

# Metrics
duration: 3min
completed: 2026-01-09
---

# Phase 9 Plan 01: Epic Command with Lifecycle Execution Summary

**Epic command now runs full lifecycle (create→dev→review→commit) for each story before moving to the next**

## Performance

- **Duration:** 3 min
- **Started:** 2026-01-09T02:26:01Z
- **Completed:** 2026-01-09T02:28:46Z
- **TDD Phases:** RED, GREEN (no REFACTOR needed)
- **Files modified:** 2

## TDD Cycle

### RED Phase

- Added comprehensive table-driven tests for epic lifecycle execution
- Test cases: 2 backlog stories, mixed statuses, done story skipped, fail-fast on failure, all done stories
- Tests verify workflows executed in order AND status updates occur
- Tests failed as expected (epic used old QueueRunner pattern)

### GREEN Phase

- Updated epic.go to use lifecycle.NewExecutor instead of QueueRunner
- For each story: execute full lifecycle, handle ErrStoryComplete for done stories
- Fail-fast on workflow failure
- Updated help text to describe full lifecycle behavior
- Removed obsolete legacy tests (tested old single-workflow behavior)

### REFACTOR Phase

- Reviewed code - no refactoring needed
- Implementation is clean and follows run.go pattern

## Task Commits

Each TDD phase committed atomically:

1. **RED: Failing tests** - `bed9b91` (test)
2. **GREEN: Implementation** - `4d2b168` (feat)

## Files Created/Modified

- `internal/cli/epic.go` - Updated to use lifecycle.Executor, full lifecycle per story
- `internal/cli/epic_test.go` - New table-driven lifecycle tests, removed obsolete legacy tests

## Decisions Made

- Removed 7 legacy tests that tested old single-workflow behavior - they were obsolete and incompatible with the new lifecycle pattern

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- Epic command now runs full lifecycle per story, consistent with run command
- Queue command is next to be updated (Phase 10) for consistency
- All existing behaviors maintained: done story skipping, fail-fast, numeric sorting

---

_Phase: 09-update-epic-command_
_Completed: 2026-01-09_
