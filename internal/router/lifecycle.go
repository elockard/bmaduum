package router

import (
	"bmaduum/internal/status"
)

// LifecycleStep represents a single step in the story lifecycle sequence.
//
// Each step contains the workflow to execute and the status to transition to
// after the workflow completes successfully. The lifecycle executor uses these
// steps to drive a story from its current status through to completion.
type LifecycleStep struct {
	// Workflow is the name of the workflow to execute for this step.
	// Must correspond to a key in the workflows configuration.
	Workflow string

	// NextStatus is the status to set after this step completes successfully.
	// The final step typically sets status to "done".
	NextStatus status.Status

	// Model is the Claude model to use for this workflow (optional).
	// If empty, the default model is used.
	Model string
}
