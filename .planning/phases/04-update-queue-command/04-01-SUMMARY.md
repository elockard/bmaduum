---
phase: 04-update-queue-command
plan: 01
subsystem: cli
tags: [cobra, status-routing, queue, yaml]

# Dependency graph
requires:
  - phase: 01-sprint-status-reader
    provides: status.Reader for reading sprint-status.yaml
  - phase: 02-workflow-router
    provides: router.GetWorkflow for status-to-workflow mapping
  - phase: 03-update-run-command
    provides: StatusReader injection pattern via App struct
provides:
  - status-based routing in queue command
  - automatic workflow selection for multiple stories
  - done story skipping (not failure) in queued batches
affects: [05-epic-command]

# Tech tracking
tech-stack:
  added: []
  patterns: [queue-skip-done-pattern]

key-files:
  created: [internal/cli/queue_test.go]
  modified:
    [
      internal/cli/queue.go,
      internal/workflow/queue.go,
      internal/output/printer.go,
    ]

key-decisions:
  - "Done stories in queue are skipped (continue to next), not terminal success"
  - "Added Skipped field to StoryResult for queue summary display"

patterns-established:
  - "Queue skip behavior: done stories don't stop queue, just skip to next"

issues-created: []

# Metrics
duration: 4min
completed: 2026-01-08
---

# Phase 4 Plan 01: Update Queue Command Summary

**Queue command now routes each story to appropriate workflow based on status, skipping done stories**

## Performance

- **Duration:** 4 min
- **Started:** 2026-01-08T20:22:33Z
- **Completed:** 2026-01-08T20:26:28Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Queue command routes each story to appropriate workflow based on sprint-status.yaml
- Done stories are skipped (not failures) - allows queuing mixed-status stories
- Queue summary now shows skipped count and marks done stories distinctly
- Updated printer to display skipped vs pending vs completed stories

## Task Commits

Each task was committed atomically:

1. **Task 1: Update queue command to use status-based routing** - `38a91d6` (feat)
2. **Task 2: Add tests for queue status-based routing** - `42de3f3` (test)

**Plan metadata:** TBD (docs: complete plan)

## Files Created/Modified

- `internal/cli/queue.go` - Updated to use RunQueueWithStatus and new description
- `internal/workflow/queue.go` - New RunQueueWithStatus with status-based routing and skip logic
- `internal/output/printer.go` - Added Skipped field to StoryResult, updated QueueSummary
- `internal/cli/queue_test.go` - Comprehensive tests for status-based routing (new)
- `internal/cli/cli_test.go` - Excluded queue from generic tests (requires sprint-status.yaml)
- `internal/output/printer_test.go` - Updated test for new pending terminology

## Decisions Made

- Done stories skipped in queue (not terminal like run command) - allows batch processing mixed status
- Added Skipped field to StoryResult to track skip reason in summary

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- Queue command updated and fully tested
- Status-based routing pattern complete for both run and queue
- Ready for Phase 5: Epic command implementation

---

_Phase: 04-update-queue-command_
_Completed: 2026-01-08_
