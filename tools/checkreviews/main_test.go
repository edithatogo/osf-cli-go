package main

import "testing"

func TestCompletedPlan(t *testing.T) {
	t.Parallel()

	complete, hasTasks := completedPlan("- [x] Task: done\n- [x] Task: also done\n")
	if !complete || !hasTasks {
		t.Fatalf("completedPlan complete=%v hasTasks=%v, want true true", complete, hasTasks)
	}

	complete, hasTasks = completedPlan("- [x] Task: done\n- [ ] Task: pending\n")
	if complete || !hasTasks {
		t.Fatalf("completedPlan complete=%v hasTasks=%v, want false true", complete, hasTasks)
	}

	complete, hasTasks = completedPlan("# No tasks\n")
	if complete || hasTasks {
		t.Fatalf("completedPlan complete=%v hasTasks=%v, want false false", complete, hasTasks)
	}
}
