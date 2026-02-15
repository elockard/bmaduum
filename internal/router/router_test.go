package router

import (
	"errors"
	"testing"

	"bmaduum/internal/manifest"
	"bmaduum/internal/status"
)

func TestGetWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		status         status.Status
		wantWorkflow   string
		wantErr        error
		wantErrMessage string
	}{
		{
			name:         "backlog status returns create-story workflow",
			status:       status.StatusBacklog,
			wantWorkflow: "create-story",
			wantErr:      nil,
		},
		{
			name:         "ready-for-dev status returns dev-story workflow",
			status:       status.StatusReadyForDev,
			wantWorkflow: "dev-story",
			wantErr:      nil,
		},
		{
			name:         "in-progress status returns dev-story workflow",
			status:       status.StatusInProgress,
			wantWorkflow: "dev-story",
			wantErr:      nil,
		},
		{
			name:         "review status returns code-review workflow",
			status:       status.StatusReview,
			wantWorkflow: "code-review",
			wantErr:      nil,
		},
		{
			name:         "done status returns ErrStoryComplete",
			status:       status.StatusDone,
			wantWorkflow: "",
			wantErr:      ErrStoryComplete,
		},
		{
			name:         "unknown status returns ErrUnknownStatus",
			status:       status.Status("invalid"),
			wantWorkflow: "",
			wantErr:      ErrUnknownStatus,
		},
		{
			name:         "empty status returns ErrUnknownStatus",
			status:       status.Status(""),
			wantWorkflow: "",
			wantErr:      ErrUnknownStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWorkflow, gotErr := GetWorkflow(tt.status)

			if gotWorkflow != tt.wantWorkflow {
				t.Errorf("GetWorkflow(%q) workflow = %q, want %q", tt.status, gotWorkflow, tt.wantWorkflow)
			}

			if tt.wantErr != nil {
				if gotErr == nil {
					t.Errorf("GetWorkflow(%q) err = nil, want %v", tt.status, tt.wantErr)
				} else if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("GetWorkflow(%q) err = %v, want %v", tt.status, gotErr, tt.wantErr)
				}
			} else if gotErr != nil {
				t.Errorf("GetWorkflow(%q) err = %v, want nil", tt.status, gotErr)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Run("ErrStoryComplete has descriptive message", func(t *testing.T) {
		if ErrStoryComplete.Error() == "" {
			t.Error("ErrStoryComplete should have a non-empty error message")
		}
	})

	t.Run("ErrUnknownStatus has descriptive message", func(t *testing.T) {
		if ErrUnknownStatus.Error() == "" {
			t.Error("ErrUnknownStatus should have a non-empty error message")
		}
	})

	t.Run("sentinel errors are distinct", func(t *testing.T) {
		if errors.Is(ErrStoryComplete, ErrUnknownStatus) {
			t.Error("ErrStoryComplete and ErrUnknownStatus should be distinct errors")
		}
	})
}

// --- Router struct tests ---

func TestNewRouter_DefaultMatches_GetWorkflow(t *testing.T) {
	r := NewRouter()

	tests := []struct {
		status       status.Status
		wantWorkflow string
		wantErr      error
	}{
		{status.StatusBacklog, "create-story", nil},
		{status.StatusReadyForDev, "dev-story", nil},
		{status.StatusInProgress, "dev-story", nil},
		{status.StatusReview, "code-review", nil},
		{status.StatusDone, "", ErrStoryComplete},
		{status.Status("invalid"), "", ErrUnknownStatus},
	}

	for _, tt := range tests {
		workflow, err := r.GetWorkflow(tt.status)
		if workflow != tt.wantWorkflow {
			t.Errorf("Router.GetWorkflow(%q) = %q, want %q", tt.status, workflow, tt.wantWorkflow)
		}
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Router.GetWorkflow(%q) err = %v, want %v", tt.status, err, tt.wantErr)
			}
		} else if err != nil {
			t.Errorf("Router.GetWorkflow(%q) unexpected err: %v", tt.status, err)
		}
	}
}

func TestNewRouter_DefaultMatches_GetLifecycle(t *testing.T) {
	r := NewRouter()

	// Verify that Router.GetLifecycle produces the same results as the package-level function
	statuses := []status.Status{
		status.StatusBacklog,
		status.StatusReadyForDev,
		status.StatusInProgress,
		status.StatusReview,
		status.StatusDone,
	}

	for _, s := range statuses {
		pkgSteps, pkgErr := GetLifecycle(s)
		routerSteps, routerErr := r.GetLifecycle(s)

		if !errors.Is(pkgErr, routerErr) {
			t.Errorf("GetLifecycle(%q): package err=%v, router err=%v", s, pkgErr, routerErr)
			continue
		}

		if len(pkgSteps) != len(routerSteps) {
			t.Errorf("GetLifecycle(%q): package len=%d, router len=%d", s, len(pkgSteps), len(routerSteps))
			continue
		}

		for i := range pkgSteps {
			if pkgSteps[i].Workflow != routerSteps[i].Workflow {
				t.Errorf("GetLifecycle(%q)[%d].Workflow: package=%q, router=%q", s, i, pkgSteps[i].Workflow, routerSteps[i].Workflow)
			}
			if pkgSteps[i].NextStatus != routerSteps[i].NextStatus {
				t.Errorf("GetLifecycle(%q)[%d].NextStatus: package=%q, router=%q", s, i, pkgSteps[i].NextStatus, routerSteps[i].NextStatus)
			}
		}
	}
}

func TestNewRouterFromManifest_GetWorkflow(t *testing.T) {
	csv := `phase,workflow,agent,command,trigger_status,next_status
3,create-story,SM,/create-story,backlog,ready-for-dev
3,dev-story,Dev,/dev-story,ready-for-dev,review
3,dev-story,Dev,/dev-story,in-progress,review
3,code-review,QA,/code-review,review,done
3,git-commit,,/git-commit,,done
`
	m, err := manifest.ReadFromString(csv)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	r := NewRouterFromManifest(m)

	tests := []struct {
		status       status.Status
		wantWorkflow string
		wantErr      error
	}{
		{status.StatusBacklog, "create-story", nil},
		{status.StatusReadyForDev, "dev-story", nil},
		{status.StatusInProgress, "dev-story", nil},
		{status.StatusReview, "code-review", nil},
		{status.StatusDone, "", ErrStoryComplete},
		{status.Status("invalid"), "", ErrUnknownStatus},
	}

	for _, tt := range tests {
		workflow, err := r.GetWorkflow(tt.status)
		if workflow != tt.wantWorkflow {
			t.Errorf("ManifestRouter.GetWorkflow(%q) = %q, want %q", tt.status, workflow, tt.wantWorkflow)
		}
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ManifestRouter.GetWorkflow(%q) err = %v, want %v", tt.status, err, tt.wantErr)
			}
		} else if err != nil {
			t.Errorf("ManifestRouter.GetWorkflow(%q) unexpected err: %v", tt.status, err)
		}
	}
}

func TestNewRouterFromManifest_GetLifecycle(t *testing.T) {
	csv := `phase,workflow,agent,command,trigger_status,next_status
3,create-story,SM,/create-story,backlog,ready-for-dev
3,dev-story,Dev,/dev-story,ready-for-dev,review
3,dev-story,Dev,/dev-story,in-progress,review
3,code-review,QA,/code-review,review,done
3,git-commit,,/git-commit,,done
`
	m, err := manifest.ReadFromString(csv)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	r := NewRouterFromManifest(m)

	tests := []struct {
		name      string
		status    status.Status
		wantSteps []LifecycleStep
		wantErr   error
	}{
		{
			name:   "backlog returns full chain",
			status: status.StatusBacklog,
			wantSteps: []LifecycleStep{
				{Workflow: "create-story", NextStatus: status.StatusReadyForDev},
				{Workflow: "dev-story", NextStatus: status.StatusReview},
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:   "ready-for-dev skips create-story",
			status: status.StatusReadyForDev,
			wantSteps: []LifecycleStep{
				{Workflow: "dev-story", NextStatus: status.StatusReview},
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:   "in-progress same as ready-for-dev",
			status: status.StatusInProgress,
			wantSteps: []LifecycleStep{
				{Workflow: "dev-story", NextStatus: status.StatusReview},
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:   "review returns code-review and git-commit",
			status: status.StatusReview,
			wantSteps: []LifecycleStep{
				{Workflow: "code-review", NextStatus: status.StatusDone},
				{Workflow: "git-commit", NextStatus: status.StatusDone},
			},
		},
		{
			name:    "done returns ErrStoryComplete",
			status:  status.StatusDone,
			wantErr: ErrStoryComplete,
		},
		{
			name:    "unknown status returns ErrUnknownStatus",
			status:  status.Status("invalid"),
			wantErr: ErrUnknownStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, err := r.GetLifecycle(tt.status)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("got err %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			if len(steps) != len(tt.wantSteps) {
				t.Fatalf("got %d steps, want %d", len(steps), len(tt.wantSteps))
			}

			for i, want := range tt.wantSteps {
				if steps[i].Workflow != want.Workflow {
					t.Errorf("step[%d].Workflow = %q, want %q", i, steps[i].Workflow, want.Workflow)
				}
				if steps[i].NextStatus != want.NextStatus {
					t.Errorf("step[%d].NextStatus = %q, want %q", i, steps[i].NextStatus, want.NextStatus)
				}
			}
		})
	}
}

func TestNewRouterFromManifest_CustomChain(t *testing.T) {
	// Test a custom manifest with different workflow names
	csv := `workflow,trigger_status,next_status
plan,backlog,ready-for-dev
implement,ready-for-dev,review
test,review,done
`
	m, err := manifest.ReadFromString(csv)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	r := NewRouterFromManifest(m)

	// GetWorkflow should use custom names
	workflow, err := r.GetWorkflow(status.StatusBacklog)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if workflow != "plan" {
		t.Errorf("GetWorkflow(backlog) = %q, want %q", workflow, "plan")
	}

	// GetLifecycle should use custom chain
	steps, err := r.GetLifecycle(status.StatusBacklog)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("got %d steps, want 3", len(steps))
	}
	if steps[0].Workflow != "plan" {
		t.Errorf("step[0].Workflow = %q, want %q", steps[0].Workflow, "plan")
	}
	if steps[1].Workflow != "implement" {
		t.Errorf("step[1].Workflow = %q, want %q", steps[1].Workflow, "implement")
	}
	if steps[2].Workflow != "test" {
		t.Errorf("step[2].Workflow = %q, want %q", steps[2].Workflow, "test")
	}
}

func TestRouter_InsertStepAfter(t *testing.T) {
	r := NewRouter()

	// Insert test-automation after code-review
	r.InsertStepAfter("code-review", "test-automation", status.StatusDone)

	// Verify the chain now has 5 steps
	steps, err := r.GetLifecycle(status.StatusBacklog)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(steps) != 5 {
		t.Fatalf("got %d steps, want 5", len(steps))
	}

	// Verify order: create-story, dev-story, code-review, test-automation, git-commit
	wantWorkflows := []string{"create-story", "dev-story", "code-review", "test-automation", "git-commit"}
	for i, want := range wantWorkflows {
		if steps[i].Workflow != want {
			t.Errorf("step[%d].Workflow = %q, want %q", i, steps[i].Workflow, want)
		}
	}

	// Verify status chain index was updated - review should now include test-automation
	reviewSteps, err := r.GetLifecycle(status.StatusReview)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(reviewSteps) != 3 {
		t.Fatalf("review: got %d steps, want 3", len(reviewSteps))
	}
	if reviewSteps[0].Workflow != "code-review" {
		t.Errorf("review step[0] = %q, want code-review", reviewSteps[0].Workflow)
	}
	if reviewSteps[1].Workflow != "test-automation" {
		t.Errorf("review step[1] = %q, want test-automation", reviewSteps[1].Workflow)
	}
	if reviewSteps[2].Workflow != "git-commit" {
		t.Errorf("review step[2] = %q, want git-commit", reviewSteps[2].Workflow)
	}
}

func TestRouter_InsertStepAfter_WorkflowNotFound(t *testing.T) {
	r := NewRouter()

	// Insert after non-existent workflow should be a no-op
	r.InsertStepAfter("nonexistent", "test-automation", status.StatusDone)

	steps, err := r.GetLifecycle(status.StatusBacklog)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(steps) != 4 {
		t.Errorf("chain should be unchanged, got %d steps, want 4", len(steps))
	}
}

func TestRouter_InsertStepAfter_DuplicateWorkflow(t *testing.T) {
	r := NewRouter()

	// First insert should work
	r.InsertStepAfter("code-review", "test-automation", status.StatusDone)
	steps, _ := r.GetLifecycle(status.StatusBacklog)
	if len(steps) != 5 {
		t.Fatalf("first insert: got %d steps, want 5", len(steps))
	}

	// Second insert of same workflow should be a no-op
	r.InsertStepAfter("code-review", "test-automation", status.StatusDone)
	steps, _ = r.GetLifecycle(status.StatusBacklog)
	if len(steps) != 5 {
		t.Errorf("duplicate insert should be no-op: got %d steps, want 5", len(steps))
	}
}

func TestRouter_InsertStepAfter_AtEnd(t *testing.T) {
	r := NewRouter()

	// Insert after the last step (git-commit)
	r.InsertStepAfter("git-commit", "deploy", status.StatusDone)

	steps, err := r.GetLifecycle(status.StatusBacklog)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(steps) != 5 {
		t.Fatalf("got %d steps, want 5", len(steps))
	}
	if steps[4].Workflow != "deploy" {
		t.Errorf("last step = %q, want deploy", steps[4].Workflow)
	}
}

func TestRouter_InsertStepAfter_PreservesGetWorkflow(t *testing.T) {
	r := NewRouter()
	r.InsertStepAfter("code-review", "test-automation", status.StatusDone)

	// GetWorkflow should still return the same single workflows
	tests := []struct {
		status       status.Status
		wantWorkflow string
	}{
		{status.StatusBacklog, "create-story"},
		{status.StatusReadyForDev, "dev-story"},
		{status.StatusInProgress, "dev-story"},
		{status.StatusReview, "code-review"},
	}

	for _, tt := range tests {
		workflow, err := r.GetWorkflow(tt.status)
		if err != nil {
			t.Errorf("GetWorkflow(%q) err: %v", tt.status, err)
		}
		if workflow != tt.wantWorkflow {
			t.Errorf("GetWorkflow(%q) = %q, want %q", tt.status, workflow, tt.wantWorkflow)
		}
	}
}

func TestNewRouterFromManifest_MatchesDefaultRouter(t *testing.T) {
	// A manifest that matches the default hardcoded routing should produce identical results
	csv := `phase,workflow,agent,command,trigger_status,next_status
3,create-story,SM,/create-story,backlog,ready-for-dev
3,dev-story,Dev,/dev-story,ready-for-dev,review
3,dev-story,Dev,/dev-story,in-progress,review
3,code-review,QA,/code-review,review,done
3,git-commit,,/git-commit,,done
`
	m, err := manifest.ReadFromString(csv)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	defaultR := NewRouter()
	manifestR := NewRouterFromManifest(m)

	statuses := []status.Status{
		status.StatusBacklog,
		status.StatusReadyForDev,
		status.StatusInProgress,
		status.StatusReview,
		status.StatusDone,
		status.Status("invalid"),
	}

	for _, s := range statuses {
		dWorkflow, dErr := defaultR.GetWorkflow(s)
		mWorkflow, mErr := manifestR.GetWorkflow(s)

		if dWorkflow != mWorkflow {
			t.Errorf("GetWorkflow(%q): default=%q, manifest=%q", s, dWorkflow, mWorkflow)
		}
		if !errors.Is(dErr, mErr) {
			t.Errorf("GetWorkflow(%q): default err=%v, manifest err=%v", s, dErr, mErr)
		}

		dSteps, dErr := defaultR.GetLifecycle(s)
		mSteps, mErr := manifestR.GetLifecycle(s)

		if !errors.Is(dErr, mErr) {
			t.Errorf("GetLifecycle(%q): default err=%v, manifest err=%v", s, dErr, mErr)
			continue
		}
		if len(dSteps) != len(mSteps) {
			t.Errorf("GetLifecycle(%q): default len=%d, manifest len=%d", s, len(dSteps), len(mSteps))
			continue
		}
		for i := range dSteps {
			if dSteps[i].Workflow != mSteps[i].Workflow || dSteps[i].NextStatus != mSteps[i].NextStatus {
				t.Errorf("GetLifecycle(%q)[%d]: default={%s,%s}, manifest={%s,%s}",
					s, i, dSteps[i].Workflow, dSteps[i].NextStatus, mSteps[i].Workflow, mSteps[i].NextStatus)
			}
		}
	}
}
