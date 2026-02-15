// Package router provides workflow routing based on story status.
//
// The router maps story status values to workflow names for single-step execution
// and provides lifecycle step sequences for multi-step execution. It serves as
// the central decision point for determining which workflow to run for a given story.
//
// Routing can be driven by hardcoded defaults ([NewRouter]) or by a BMAD v6
// workflow manifest ([NewRouterFromManifest]) for dynamic workflow discovery.
//
// Key types:
//   - [Router] - Configurable workflow router (hardcoded or manifest-driven)
//   - [LifecycleStep] - A single step in a lifecycle sequence
//
// Package-level functions [GetWorkflow] and [GetLifecycle] use the default
// hardcoded router for backward compatibility.
package router

import (
	"errors"

	"bmaduum/internal/manifest"
	"bmaduum/internal/status"
)

// Sentinel errors for workflow routing.
var (
	// ErrStoryComplete is a sentinel error indicating the story has status "done"
	// and no workflow is needed. Callers should skip the story rather than treat
	// this as a failure condition.
	ErrStoryComplete = errors.New("story is complete, no workflow needed")

	// ErrUnknownStatus is a sentinel error indicating the status value is not
	// recognized. Callers should report this as an error, as it likely indicates
	// a typo in the sprint-status.yaml file.
	ErrUnknownStatus = errors.New("unknown status value")
)

// chainStep is an internal representation of a step in the workflow chain.
type chainStep struct {
	Workflow   string
	NextStatus status.Status
}

// Router routes story statuses to workflows.
//
// Create with [NewRouter] for hardcoded defaults or [NewRouterFromManifest]
// for manifest-driven routing. The router supports two modes of operation:
//   - Single-step: [Router.GetWorkflow] returns one workflow for a status
//   - Multi-step: [Router.GetLifecycle] returns all remaining steps to completion
type Router struct {
	// statusWorkflow maps trigger status → workflow name for single-step routing.
	statusWorkflow map[status.Status]string

	// chain is the ordered workflow chain for lifecycle execution.
	chain []chainStep

	// statusChainIndex maps trigger status → index into chain where execution starts.
	statusChainIndex map[status.Status]int
}

// NewRouter creates a [Router] with the default hardcoded routing rules.
//
// The default chain is: create-story → dev-story → code-review → git-commit.
// Status mappings are:
//   - backlog → create-story
//   - ready-for-dev, in-progress → dev-story
//   - review → code-review
//   - done → [ErrStoryComplete]
func NewRouter() *Router {
	return &Router{
		statusWorkflow: map[status.Status]string{
			status.StatusBacklog:     "create-story",
			status.StatusReadyForDev: "dev-story",
			status.StatusInProgress:  "dev-story",
			status.StatusReview:      "code-review",
		},
		chain: []chainStep{
			{Workflow: "create-story", NextStatus: status.StatusReadyForDev},
			{Workflow: "dev-story", NextStatus: status.StatusReview},
			{Workflow: "code-review", NextStatus: status.StatusDone},
			{Workflow: "git-commit", NextStatus: status.StatusDone},
		},
		statusChainIndex: map[status.Status]int{
			status.StatusBacklog:     0,
			status.StatusReadyForDev: 1,
			status.StatusInProgress:  1,
			status.StatusReview:      2,
		},
	}
}

// NewRouterFromManifest creates a [Router] from a BMAD v6 workflow manifest.
//
// The manifest entries define:
//   - The workflow chain order (from entry order in the manifest)
//   - Status-to-workflow mappings (from trigger_status fields)
//   - Status transitions (from next_status fields)
//
// Entries without a trigger_status are included in the lifecycle chain but
// are not directly triggerable by status (e.g., git-commit).
func NewRouterFromManifest(m *manifest.Manifest) *Router {
	r := &Router{
		statusWorkflow:   make(map[status.Status]string),
		statusChainIndex: make(map[status.Status]int),
	}

	// Build the chain from unique workflows in manifest order
	seen := make(map[string]bool)
	for _, entry := range m.Entries {
		if seen[entry.Workflow] {
			// Already added this workflow to the chain; just add trigger status mapping
			if entry.TriggerStatus != "" {
				s := status.Status(entry.TriggerStatus)
				r.statusWorkflow[s] = entry.Workflow
				// Find the chain index for this workflow
				for i, step := range r.chain {
					if step.Workflow == entry.Workflow {
						r.statusChainIndex[s] = i
						break
					}
				}
			}
			continue
		}
		seen[entry.Workflow] = true

		// Add to chain
		r.chain = append(r.chain, chainStep{
			Workflow:   entry.Workflow,
			NextStatus: status.Status(entry.NextStatus),
		})

		// Add trigger status mapping
		if entry.TriggerStatus != "" {
			s := status.Status(entry.TriggerStatus)
			r.statusWorkflow[s] = entry.Workflow
			r.statusChainIndex[s] = len(r.chain) - 1
		}
	}

	return r
}

// GetWorkflow returns the single workflow name for the given story status.
//
// Returns [ErrStoryComplete] for done stories (caller should skip, not fail).
// Returns [ErrUnknownStatus] for unrecognized status values.
func (r *Router) GetWorkflow(s status.Status) (string, error) {
	if s == status.StatusDone {
		return "", ErrStoryComplete
	}

	workflow, ok := r.statusWorkflow[s]
	if !ok {
		return "", ErrUnknownStatus
	}
	return workflow, nil
}

// GetLifecycle returns the complete sequence of lifecycle steps from the given
// status through to completion.
//
// Returns [ErrStoryComplete] for done stories (caller should skip, not fail).
// Returns [ErrUnknownStatus] for unrecognized status values.
func (r *Router) GetLifecycle(s status.Status) ([]LifecycleStep, error) {
	if s == status.StatusDone {
		return nil, ErrStoryComplete
	}

	startIdx, ok := r.statusChainIndex[s]
	if !ok {
		return nil, ErrUnknownStatus
	}

	// Build lifecycle steps from the chain starting at startIdx
	remaining := r.chain[startIdx:]
	steps := make([]LifecycleStep, len(remaining))
	for i, cs := range remaining {
		steps[i] = LifecycleStep{
			Workflow:   cs.Workflow,
			NextStatus: cs.NextStatus,
		}
	}

	return steps, nil
}

// defaultRouter is the package-level router used by backward-compatible functions.
var defaultRouter = NewRouter()

// GetWorkflow returns the single workflow name for the given story status.
//
// This package-level function uses the default hardcoded router for backward
// compatibility. For manifest-driven routing, create a [Router] with
// [NewRouterFromManifest] and call its GetWorkflow method.
//
// The mapping is:
//   - backlog -> "create-story"
//   - ready-for-dev, in-progress -> "dev-story"
//   - review -> "code-review"
//   - done -> [ErrStoryComplete]
//
// Returns [ErrStoryComplete] for done stories (caller should skip, not fail).
// Returns [ErrUnknownStatus] for unrecognized status values (likely YAML typo).
//
// See [status.Status] for valid status values.
func GetWorkflow(s status.Status) (string, error) {
	return defaultRouter.GetWorkflow(s)
}

// GetLifecycle returns the complete sequence of lifecycle steps from the given
// status through to "done".
//
// This package-level function uses the default hardcoded router for backward
// compatibility. For manifest-driven routing, create a [Router] with
// [NewRouterFromManifest] and call its GetLifecycle method.
//
// The sequences are:
//   - backlog: create-story -> dev-story -> code-review -> git-commit -> done
//   - ready-for-dev, in-progress: dev-story -> code-review -> git-commit -> done
//   - review: code-review -> git-commit -> done
//   - done: [ErrStoryComplete]
//
// Returns [ErrStoryComplete] for done stories (caller should skip, not fail).
// Returns [ErrUnknownStatus] for unrecognized status values (likely YAML typo).
//
// See [status.Status] for valid status values.
func GetLifecycle(s status.Status) ([]LifecycleStep, error) {
	return defaultRouter.GetLifecycle(s)
}
