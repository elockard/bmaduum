// Package workflow provides workflow orchestration for bmaduum.
//
// Workflows execute Claude CLI with configured prompts to automate development tasks.
// The package supports single workflow execution and raw prompt execution.
//
// Key types:
//   - [Runner] orchestrates individual Claude executions with output formatting
//
// The [Runner] requires a [claude.Executor] for spawning Claude CLI processes.
// Workflow prompts are configured in the config package and support Go template
// expansion with story keys.
package workflow
