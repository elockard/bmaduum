# Batch Processing Recipes

Multi-story and epic processing patterns.

---

## Process Multiple Stories

Run full lifecycle for each story in sequence.

```bash
bmaduum queue AUTH-042 AUTH-043 AUTH-044
```

Each story runs through its complete lifecycle. Skips stories already at `done`. Stops on first failure.

**Example output:**

```
Queue: 3 stories [AUTH-042, AUTH-043, AUTH-044]

[1/3] AUTH-042
  ... workflow output ...

[2/3] AUTH-043
  ... workflow output ...

Summary:
  AUTH-042  ✓  1m 23s
  AUTH-043  ✓  2m 45s
  AUTH-044  ○  skipped (done)
```

---

## Preview Batch Execution

See the full execution plan before running.

```bash
bmaduum queue --dry-run AUTH-042 AUTH-043 AUTH-044
```

**Example output:**

```
Dry run for 3 stories:

Story AUTH-042:
  1. dev-story -> review
  2. code-review -> done
  3. git-commit -> done

Story AUTH-043:
  (already complete)

Story AUTH-044:
  1. create-story -> ready-for-dev
  2. dev-story -> review
  3. code-review -> done
  4. git-commit -> done

Total: 7 workflows across 2 stories (1 already complete)
```

---

## Process an Entire Epic

Run all stories matching an epic pattern.

```bash
bmaduum epic 05
```

Finds stories matching `05-*` (e.g., `05-01-auth`, `05-02-dashboard`), sorts by story number, runs each through its complete lifecycle.

---

## Process Multiple Epics

Run multiple epics in sequence.

```bash
bmaduum epic 02 04 06
```

Processes epic 02, then 04, then 06. Each epic's stories are run to completion before moving to the next epic.

---

## Preview Epic Execution

See what stories would be processed and their workflows.

```bash
# Single epic
bmaduum epic --dry-run 05

# Multiple epics
bmaduum epic --dry-run 02 04 06
```

Shows all matching stories and their remaining lifecycle steps.

---

## Handle Mixed-Status Batches

When processing stories with different statuses, each runs only its remaining workflows.

```bash
# Stories at different statuses:
# AUTH-042: ready-for-dev
# AUTH-043: review
# AUTH-044: done

bmaduum queue AUTH-042 AUTH-043 AUTH-044
```

**Result:**

- AUTH-042: Runs dev-story -> code-review -> git-commit
- AUTH-043: Runs code-review -> git-commit
- AUTH-044: Skipped (already done)

Completed stories are always skipped, not re-run.
