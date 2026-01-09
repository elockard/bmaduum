# Roadmap: BMAD Automate

## Milestones

- [v1.0 Status-Based Workflow Routing](milestones/v1.0-ROADMAP.md) (Phases 1-5) — SHIPPED 2026-01-08
- [v1.1 Full Story Lifecycle](milestones/v1.1-ROADMAP.md) (Phases 6-13) — SHIPPED 2026-01-09
- 🚧 **v1.2 Documentation** - Phases 14-19 (in progress)

## Completed Milestones

<details>
<summary>v1.0 Status-Based Workflow Routing (Phases 1-5) — SHIPPED 2026-01-08</summary>

**Delivered:** Automatic workflow routing based on sprint-status.yaml, eliminating manual workflow selection.

- [x] Phase 1: Sprint Status Reader (1/1 plans) — completed 2026-01-08
- [x] Phase 2: Workflow Router (1/1 plans) — completed 2026-01-08
- [x] Phase 3: Update Run Command (1/1 plans) — completed 2026-01-08
- [x] Phase 4: Update Queue Command (1/1 plans) — completed 2026-01-08
- [x] Phase 5: Epic Command (1/1 plans) — completed 2026-01-08

</details>

<details>
<summary>v1.1 Full Story Lifecycle (Phases 6-13) — SHIPPED 2026-01-09</summary>

**Delivered:** Complete story lifecycle execution (create→dev→review→commit) with error recovery, dry-run mode, and step progress visibility.

- [x] Phase 6: Lifecycle Definition (1/1 plans) — completed 2026-01-08
- [x] Phase 7: Story Lifecycle Executor (2/2 plans) — completed 2026-01-09
- [x] Phase 8: Update Run Command (1/1 plans) — completed 2026-01-09
- [x] Phase 9: Update Epic Command (1/1 plans) — completed 2026-01-09
- [x] Phase 10: Update Queue Command (1/1 plans) — completed 2026-01-09
- [x] Phase 11: Error Recovery & Resume (1/1 plans) — completed 2026-01-09
- [x] Phase 12: Dry Run Mode (2/2 plans) — completed 2026-01-09
- [x] Phase 13: Enhanced Progress UI (1/1 plans) — completed 2026-01-09

</details>

## Progress

| Phase                       | Milestone | Plans Complete | Status   | Completed  |
| --------------------------- | --------- | -------------- | -------- | ---------- |
| 1. Sprint Status Reader     | v1.0      | 1/1            | Complete | 2026-01-08 |
| 2. Workflow Router          | v1.0      | 1/1            | Complete | 2026-01-08 |
| 3. Update Run Command       | v1.0      | 1/1            | Complete | 2026-01-08 |
| 4. Update Queue Command     | v1.0      | 1/1            | Complete | 2026-01-08 |
| 5. Epic Command             | v1.0      | 1/1            | Complete | 2026-01-08 |
| 6. Lifecycle Definition     | v1.1      | 1/1            | Complete | 2026-01-08 |
| 7. Story Lifecycle Executor | v1.1      | 2/2            | Complete | 2026-01-09 |
| 8. Update Run Command       | v1.1      | 1/1            | Complete | 2026-01-09 |
| 9. Update Epic Command      | v1.1      | 1/1            | Complete | 2026-01-09 |
| 10. Update Queue Command    | v1.1      | 1/1            | Complete | 2026-01-09 |
| 11. Error Recovery & Resume | v1.1      | 1/1            | Complete | 2026-01-09 |
| 12. Dry Run Mode            | v1.1      | 2/2            | Complete | 2026-01-09 |
| 13. Enhanced Progress UI    | v1.1      | 1/1            | Complete | 2026-01-09 |

### 🚧 v1.2 Documentation (In Progress)

**Milestone Goal:** Comprehensive documentation for open-sourced project — go doc comments throughout codebase, updated docs/, contribution guides, and API examples.

#### Phase 14: GoDoc Core Packages

**Goal**: Add comprehensive go doc comments to `internal/claude`, `internal/lifecycle`, `internal/workflow`
**Depends on**: Previous milestone complete
**Research**: Unlikely (standard Go documentation patterns)
**Plans**: 3

Plans:

- [x] 14-01: internal/claude package documentation
- [x] 14-02: internal/lifecycle package documentation
- [x] 14-03: internal/workflow package documentation

#### Phase 15: GoDoc Supporting Packages

**Goal**: Add doc comments to `internal/cli`, `internal/router`, `internal/status`, `internal/state`, `internal/output`, `internal/config`
**Depends on**: Phase 14
**Research**: Unlikely (standard Go documentation patterns)
**Plans**: TBD

Plans:

- [ ] 15-01: TBD

#### Phase 16: Package Documentation

**Goal**: Add package-level doc.go files with overviews and examples for each package
**Depends on**: Phase 15
**Research**: Unlikely (standard Go documentation patterns)
**Plans**: TBD

Plans:

- [ ] 16-01: TBD

#### Phase 17: Update Docs Folder

**Goal**: Update existing docs (README, USER_GUIDE, CLI_REFERENCE, ARCHITECTURE, PACKAGES, DEVELOPMENT) for accuracy and completeness
**Depends on**: Phase 16
**Research**: Unlikely (internal content)
**Plans**: TBD

Plans:

- [ ] 17-01: TBD

#### Phase 18: Contribution Guide

**Goal**: Add CONTRIBUTING.md, CODE_OF_CONDUCT.md, and update root README for open source
**Depends on**: Phase 17
**Research**: Unlikely (standard open source patterns)
**Plans**: TBD

Plans:

- [ ] 18-01: TBD

#### Phase 19: API Examples

**Goal**: Add runnable examples and usage patterns in docs/examples/ folder
**Depends on**: Phase 18
**Research**: Unlikely (internal patterns)
**Plans**: TBD

Plans:

- [ ] 19-01: TBD

## Progress (v1.2)

| Phase                         | Milestone | Plans Complete | Status      | Completed  |
| ----------------------------- | --------- | -------------- | ----------- | ---------- |
| 14. GoDoc Core Packages       | v1.2      | 3/3            | Complete    | 2026-01-09 |
| 15. GoDoc Supporting Packages | v1.2      | 0/?            | Not started | -          |
| 16. Package Documentation     | v1.2      | 0/?            | Not started | -          |
| 17. Update Docs Folder        | v1.2      | 0/?            | Not started | -          |
| 18. Contribution Guide        | v1.2      | 0/?            | Not started | -          |
| 19. API Examples              | v1.2      | 0/?            | Not started | -          |
