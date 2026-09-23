package beads

import (
	"context"
	"testing"
	"time"
)

func TestNativeDoltStoreParentProjectionIncludesEphemeralChild(t *testing.T) {
	store := newNativeDoltStoreForTest(newNativeDoltMemStorage())
	result, err := store.ApplyGraphPlanWithStorage(t.Context(), &GraphApplyPlan{
		CommitMessage: "create ephemeral reparent fixture",
		Nodes: []GraphApplyNode{
			{Key: "old", Title: "Old parent"},
			{Key: "new", Title: "New parent"},
			{Key: "child", Title: "Child", ParentKey: "old"},
		},
	}, StorageEphemeral)
	if err != nil {
		t.Fatalf("ApplyGraphPlanWithStorage: %v", err)
	}
	childID, oldParentID, newParentID := result.IDs["child"], result.IDs["old"], result.IDs["new"]
	child, err := store.Get(childID)
	if err != nil {
		t.Fatalf("Get child: %v", err)
	}
	if !child.Ephemeral || child.ParentID != oldParentID {
		t.Fatalf("created child = %+v, want ephemeral under old parent", child)
	}
	if err := store.Update(childID, UpdateOpts{ParentID: &newParentID}); err != nil {
		t.Fatalf("Update parent: %v", err)
	}
	newChildren, err := store.Children(newParentID, WithBothTiers)
	if err != nil {
		t.Fatalf("Children with both tiers: %v", err)
	}
	if !beadSliceContains(newChildren, childID) {
		t.Fatalf("new parent's both-tier children = %v, want ephemeral child %q", newChildren, childID)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	if err := store.WaitForParentProjection(ctx, childID, oldParentID, newParentID); err != nil {
		t.Fatalf("WaitForParentProjection: %v", err)
	}
}
