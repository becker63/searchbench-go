package codegraph

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestMinCallHopsAcrossFiles_twoFiles(t *testing.T) {
	a := domain.CanonicalizePath("a.go")
	b := domain.CanonicalizePath("b.go")
	c := domain.CanonicalizePath("c.go")
	store := NewStore()

	fa := NewFileNode(NodeID("file:a.go"), a)
	fb := NewFileNode(NodeID("file:b.go"), b)
	fc := NewFileNode(NodeID("file:c.go"), c)
	x := NewFunctionNode(NodeID("fn:a.go:X:1"), a, "X", 1, 2)
	y := NewFunctionNode(NodeID("fn:c.go:Y:1"), c, "Y", 1, 2)
	z := NewFunctionNode(NodeID("fn:b.go:Z:1"), b, "Z", 1, 2)

	for _, err := range []error{
		store.AddNode(fa),
		store.AddNode(fb),
		store.AddNode(fc),
		store.AddNode(x),
		store.AddNode(y),
		store.AddNode(z),
		store.AddEdge(NewEdge(fa.ID, x.ID, EdgeDefines)),
		store.AddEdge(NewEdge(fc.ID, y.ID, EdgeDefines)),
		store.AddEdge(NewEdge(fb.ID, z.ID, EdgeDefines)),
		store.AddEdge(NewEdge(x.ID, y.ID, EdgeCalls)),
		store.AddEdge(NewEdge(y.ID, z.ID, EdgeCalls)),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	g, err := store.Build()
	if err != nil {
		t.Fatal(err)
	}

	hops, ok := MinCallHopsAcrossFiles(g, []domain.RepoRelPath{a}, []domain.RepoRelPath{b})
	if !ok {
		t.Fatal("expected hops across files")
	}
	if hops != 2 {
		t.Fatalf("expected 2 hops X→Y→Z, got %d", hops)
	}
}
