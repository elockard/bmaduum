# Roadmap: BMAD Automate - Status-Based Workflow Routing

## Overview

Transform BMAD Automate from manual workflow selection to automatic status-based routing. The CLI will read story status from sprint-status.yaml and automatically select the correct workflow (create-story, dev-story, or code-review). Culminates in a new `epic` command for batch execution.

## Domain Expertise

None

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Sprint Status Reader** - Parse sprint-status.yaml and extract story statuses
- [ ] **Phase 2: Workflow Router** - Map story status to correct workflow
- [ ] **Phase 3: Update Run Command** - Apply status-based routing to run command
- [ ] **Phase 4: Update Queue Command** - Apply status-based routing to queue command
- [ ] **Phase 5: Epic Command** - New command to batch-run all epic stories with fail-fast

## Phase Details

### Phase 1: Sprint Status Reader

**Goal**: Create a package that reads `_bmad-output/implementation-artifacts/sprint-status.yaml` and returns story status by key
**Depends on**: Nothing (first phase)
**Research**: Unlikely (YAML parsing in Go, established patterns)
**Plans**: TBD

### Phase 2: Workflow Router

**Goal**: Create routing logic that maps status values to workflow names
**Depends on**: Phase 1
**Research**: Unlikely (internal logic, status-to-workflow mapping table)
**Plans**: TBD

### Phase 3: Update Run Command

**Goal**: Modify `run` command to use status-based routing instead of explicit workflow flag
**Depends on**: Phase 2
**Research**: Unlikely (modifying existing command, codebase patterns known)
**Plans**: TBD

### Phase 4: Update Queue Command

**Goal**: Modify `queue` command to use status-based routing for each story
**Depends on**: Phase 3
**Research**: Unlikely (same pattern as run command)
**Plans**: TBD

### Phase 5: Epic Command

**Goal**: New `bmad-automate epic <epic-id>` command that runs all non-done stories in numeric order, stopping on first failure
**Depends on**: Phase 4
**Research**: Unlikely (follows existing command patterns, reuses router)
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase                   | Plans Complete | Status      | Completed  |
| ----------------------- | -------------- | ----------- | ---------- |
| 1. Sprint Status Reader | 1/1            | Complete    | 2026-01-08 |
| 2. Workflow Router      | 0/TBD          | Not started | -          |
| 3. Update Run Command   | 0/TBD          | Not started | -          |
| 4. Update Queue Command | 0/TBD          | Not started | -          |
| 5. Epic Command         | 0/TBD          | Not started | -          |
