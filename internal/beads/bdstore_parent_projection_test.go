package beads

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestBdStoreParentProjectionIncludesEphemeralChild(t *testing.T) {
	runner := func(_, _ string, args ...string) ([]byte, error) {
		command := strings.Join(args, " ")
		switch command {
		case "show --json bd-child":
			return nil, fmt.Errorf("issue bd-child not found")
		case "query --json ephemeral=true AND id=bd-child --all --limit 1":
			return []byte(`[{"id":"bd-child","title":"child","status":"open","issue_type":"task","parent":"bd-new","ephemeral":true}]`), nil
		case "list --json --include-infra --include-gates --include-templates --limit 0 --parent bd-old",
			"list --json --include-infra --include-gates --include-templates --limit 0 --parent bd-new",
			"list --json --include-infra --include-gates --limit 0 --parent bd-old",
			"list --json --include-infra --include-gates --limit 0 --parent bd-new",
			"query --json ephemeral=true AND parent=bd-old --limit 0":
			return []byte(`[]`), nil
		case "query --json ephemeral=true AND parent=bd-new --limit 0":
			return []byte(`[{"id":"bd-child","title":"child","status":"open","issue_type":"task","parent":"bd-new","ephemeral":true}]`), nil
		default:
			return nil, fmt.Errorf("unexpected bd command: %s", command)
		}
	}
	store := NewBdStore("/city", runner)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	if err := store.WaitForParentProjection(ctx, "bd-child", "bd-old", "bd-new"); err != nil {
		t.Fatalf("WaitForParentProjection of ephemeral child: %v", err)
	}
}
