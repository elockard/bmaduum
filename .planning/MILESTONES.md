# Project Milestones: BMAD Automate

## v1.2 Documentation (Shipped: 2026-01-09)

**Delivered:** Comprehensive documentation for open-sourced project — go doc comments, package examples, updated docs folder, and CLI cookbook.

**Phases completed:** 14-18 (13 plans total)

**Key accomplishments:**

- Comprehensive go doc comments for all internal packages (claude, lifecycle, workflow, cli, router, status, state, output, config)
- 10 runnable Example functions in doc_test.go files for godoc and testing
- Package-level doc.go files with overviews for each package
- Updated docs folder — ARCHITECTURE.md, README.md, USER_GUIDE.md, CLI_REFERENCE.md, DEVELOPMENT.md, PACKAGES.md reflect v1.1
- CLI Cookbook with 14 recipe-style examples in docs/examples/

**Stats:**

- 68 files created/modified (+6,922 lines, -325 lines)
- 8,448 lines of Go total (up from 6,418)
- 5 phases, 13 plans
- Same-day completion (2026-01-09)

**Git range:** `docs(14-01)` → `docs(18-01)`

**What's next:** Ready for open source release

---

## v1.1 Full Story Lifecycle (Shipped: 2026-01-09)

**Delivered:** Complete story lifecycle execution (create→dev→review→commit) with error recovery, dry-run mode, and step progress visibility.

**Phases completed:** 6-13 (11 plans total)

**Key accomplishments:**

- Full story lifecycle execution from any status to done with auto-status updates
- Lifecycle orchestration via new lifecycle.Executor with interface-based DI
- State persistence for error recovery (`.bmad-state.json`) enabling resume capability
- Dry-run mode (`--dry-run`) previews workflow sequence without execution
- Step progress visibility with real-time callbacks during lifecycle execution
- Commands (run, queue, epic) all use lifecycle executor for consistency

**Stats:**

- 48 files created/modified (+7,984 lines, -534 lines)
- 6,418 lines of Go total (up from 4,951)
- 8 phases, 11 plans
- Same-day completion (2026-01-08 to 2026-01-09, ~12 hours)

**Git range:** `feat(06-01)` → `feat(13-01)`

**What's next:** TBD - milestone complete

---

## v1.0 Status-Based Workflow Routing (Shipped: 2026-01-08)

**Delivered:** Automatic workflow routing based on sprint-status.yaml, eliminating manual workflow selection.

**Phases completed:** 1-5 (5 plans total)

**Key accomplishments:**

- Sprint status reader package parsing YAML with Status type and validation
- Workflow router mapping status values to workflow names (backlog→create-story, ready-for-dev/in-progress→dev-story, review→code-review)
- Run command with automatic status-based workflow routing
- Queue command with status-based routing and done-story skipping
- New epic command for batch-running all stories in an epic with numeric sorting

**Stats:**

- 29 files created/modified (+2,636 lines, -249 lines)
- 4,951 lines of Go total
- 5 phases, 5 plans, 10 tasks
- Same-day completion (2026-01-08)

**Git range:** `docs(01)` → `docs(05-01)`

**What's next:** TBD - milestone complete

---
