package entity

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/timeutils"
)

func TestTreeNodeContainsItselfAndItsDescendantsOnly(t *testing.T) {
	collection := sampleTree().Collections()[0]
	cases := map[string]bool{"c1": true, "f1": true, "r1": true, "r2": true, "error:x": false, "missing": false}
	for id, want := range cases {
		if got := collection.Contains(id); got != want {
			t.Errorf("Contains(%s) = %v, want %v", id, got, want)
		}
	}
	if collection.Children[0].Contains("r2") {
		t.Error("a folder does not contain its siblings")
	}
	if !collection.IsCollection() || collection.Children[0].IsCollection() {
		t.Error("only collections are collections")
	}
}

func TestTreeNodeHoldsEntryIgnoringCase(t *testing.T) {
	collection := sampleTree().Collections()[0]
	if !collection.HoldsEntry("PING.yaml") || !collection.HoldsEntry("notes") || collection.HoldsEntry("pong.yaml") {
		t.Error("HoldsEntry must match the entries ignoring case")
	}
}

func TestTreeNodeOrderInsertingAndMovingClampTheIndex(t *testing.T) {
	collection := sampleTree().Collections()[0]
	cases := map[string]struct {
		got  []string
		want []string
	}{
		"insert first":   {got: collection.OrderInserting("new.yaml", 0), want: []string{"new.yaml", "invoices", "broken.yaml", "ping.yaml"}},
		"insert middle":  {got: collection.OrderInserting("new.yaml", 1), want: []string{"invoices", "new.yaml", "broken.yaml", "ping.yaml"}},
		"insert last":    {got: collection.OrderInserting("new.yaml", 3), want: []string{"invoices", "broken.yaml", "ping.yaml", "new.yaml"}},
		"insert beyond":  {got: collection.OrderInserting("new.yaml", 4), want: []string{"invoices", "broken.yaml", "ping.yaml", "new.yaml"}},
		"insert below":   {got: collection.OrderInserting("new.yaml", -1), want: []string{"new.yaml", "invoices", "broken.yaml", "ping.yaml"}},
		"move first":     {got: collection.OrderMoving("ping.yaml", 0), want: []string{"ping.yaml", "invoices", "broken.yaml"}},
		"move middle":    {got: collection.OrderMoving("invoices", 1), want: []string{"broken.yaml", "invoices", "ping.yaml"}},
		"move last":      {got: collection.OrderMoving("invoices", 2), want: []string{"broken.yaml", "ping.yaml", "invoices"}},
		"move beyond":    {got: collection.OrderMoving("invoices", 3), want: []string{"broken.yaml", "ping.yaml", "invoices"}},
		"move unchanged": {got: collection.OrderMoving("broken.yaml", 1), want: []string{"invoices", "broken.yaml", "ping.yaml"}},
	}
	for name, tc := range cases {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("%s = %v, want %v", name, tc.got, tc.want)
		}
	}
}

func TestNewTreeMoveRewritesOneOrderForAReorderAndBothOtherwise(t *testing.T) {
	tree := sampleTree()
	collection := tree.Collections()[0]
	folder := collection.Children[0]
	ping := collection.Children[2]

	reorder := NewTreeMove(ping, collection, collection, "ping.yaml", 0)
	if !reorder.IsReorder() || reorder.SourceOrder != nil || !reflect.DeepEqual(reorder.TargetOrder, []string{"ping.yaml", "invoices", "broken.yaml"}) {
		t.Errorf("reorder = %+v", reorder)
	}
	moved := NewTreeMove(ping, collection, folder, "ping-2.yaml", 0)
	want := TreeMove{
		ID: "r2", FileName: "ping-2.yaml", SourceParentID: "c1", TargetParentID: "f1",
		SourceOrder: []string{"invoices", "broken.yaml"}, TargetOrder: []string{"ping-2.yaml", "create.yaml"},
	}
	if moved.IsReorder() || !reflect.DeepEqual(moved, want) {
		t.Errorf("move = %+v, want %+v", moved, want)
	}
	collectionMove := NewTreeMove(collection, tree.Root, tree.Root, "billing", 0)
	if !collectionMove.IsReorder() || !reflect.DeepEqual(collectionMove.TargetOrder, []string{"billing"}) {
		t.Errorf("collection move = %+v", collectionMove)
	}
}

func TestRecordedMoveCoversBothSidesForTwoSeconds(t *testing.T) {
	now := time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)
	timeutils.SetCurrentTimeProvider(func() time.Time { return now })
	t.Cleanup(func() { timeutils.SetCurrentTimeProvider(time.Now) })
	source := filepath.Join("ws", "billing", "invoices")
	target := filepath.Join("ws", "users", "invoices")
	move := NewRecordedMove(source, target)

	for _, path := range []string{source, target, filepath.Join(target, "create.yaml"), filepath.Join(source, "drafts")} {
		if !move.Covers(path) {
			t.Errorf("Covers(%s) = false", path)
		}
	}
	for _, path := range []string{filepath.Join("ws", "billing"), filepath.Join("ws", "billing", "invoices-2"), filepath.Join("ws", "users", "collection.yaml")} {
		if move.Covers(path) {
			t.Errorf("Covers(%s) = true", path)
		}
	}
	now = now.Add(2 * time.Second)
	if !move.IsRecent() || !move.Covers(target) {
		t.Error("a move is still recent after exactly two seconds")
	}
	now = now.Add(time.Nanosecond)
	if move.IsRecent() || move.Covers(target) {
		t.Error("a move older than two seconds is no longer covered")
	}
}

func TestTreeNodeKindHelpersForMoves(t *testing.T) {
	folder := TreeNode{ID: "f", Kind: itemkind.Folder, Children: []TreeNode{{ID: "error:x", Kind: itemkind.Error, Children: []TreeNode{{ID: "hidden", Kind: itemkind.Request}}}}}
	if folder.Contains("hidden") {
		t.Error("error nodes do not hold items")
	}
}

func TestTopLevelParentAndRenamesOnMove(t *testing.T) {
	if !IsTopLevelParent("") || IsTopLevelParent("c1") {
		t.Fatal("IsTopLevelParent")
	}
	item := TreeNode{ID: "r1", FileName: "status.yaml"}
	source := TreeNode{ID: "c1", Entries: []string{"status.yaml"}}
	target := TreeNode{ID: "c2", Entries: []string{"status.yaml"}}
	empty := TreeNode{ID: "c3"}
	if !target.RenamesOnMove(item, source) || source.RenamesOnMove(item, source) || empty.RenamesOnMove(item, source) {
		t.Fatal("RenamesOnMove")
	}
}
