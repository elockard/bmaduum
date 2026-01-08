---
phase: 05-epic-command
plan: 01
subsystem: cli
tags: [cobra, status-reader, queue-runner, batch-execution]

# Dependency graph
requires:
  - phase: 04-update-queue-command
    provides: Queue runner with status-based routing and done-skip behavior
provides:
  - GetEpicStories method for filtering/sorting stories by epic ID
  - Epic command for batch-running all stories in an epic
affects: [future-commands, epic-management]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Epic ID prefix matching with numeric story sorting
    - Command reuses QueueRunner for status-based routing

key-files:
  created:
    - internal/cli/epic.go
    - internal/cli/epic_test.go
  modified:
    - internal/status/reader.go
    - internal/status/reader_test.go
    - internal/cli/root.go

key-decisions:
  - "Reuse QueueRunner for epic execution (inherits skip-done and fail-fast)"
  - "Story keys parsed as {epicID}-{N}-{description} with numeric N sorting"

patterns-established:
  - "Epic story discovery: prefix match + numeric second segment"

issues-created: []

# Metrics
duration: 2min
completed: 2026-01-08
---

# Phase 5 Plan 01: Epic Command Summary

**Epic command that batch-runs all stories for an epic ID using status-based routing with numeric sorting and fail-fast**

## Performance

- **Duration:** 2 min
- **Started:** 2026-01-08T20:30:03Z
- **Completed:** 2026-01-08T20:32:57Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added GetEpicStories method to status.Reader (filters by prefix, sorts numerically)
- Created epic command that finds all stories for an epic and runs them via queue runner
- Inherited queue behavior: status-based routing, skip done stories, fail-fast on error

## Task Commits

Each task was committed atomically:

1. **Task 1: Add GetEpicStories method and epic command** - `6b99a6c` (feat)
2. **Task 2: Add comprehensive tests for epic command** - `2511f25` (test)

## Files Created/Modified

- `internal/status/reader.go` - Added GetEpicStories method
- `internal/status/reader_test.go` - Added tests for GetEpicStories
- `internal/cli/epic.go` - New epic command
- `internal/cli/epic_test.go` - Comprehensive epic command tests
- `internal/cli/root.go` - Registered epic command

## Decisions Made

- Reused QueueRunner for epic execution - inherits all status-based routing, done-skip, and fail-fast behavior without duplication
- Story key pattern: `{epicID}-{N}-{description}` where N is parsed as integer for numeric sorting (2 before 10)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- All 5 phases of milestone complete
- Status-based workflow routing fully implemented
- Epic command available for batch execution

---

_Phase: 05-epic-command_
_Completed: 2026-01-08_
