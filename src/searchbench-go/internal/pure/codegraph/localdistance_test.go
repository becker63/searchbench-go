package codegraph_test

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestMinCallHopsAcrossFiles_twoFiles(t *testing.T) {
	a := domain.CanonicalizePath("a.go")
	b := domain.CanonicalizePath("b.go")
	c := domain.CanonicalizePath("c.go")
	store := codegraph.NewStore()

	fa := codegraph.NewFileNode(codegraph.NodeID("file:a.go"), a)
	fb := codegraph.NewFileNode(codegraph.NodeID("file:b.go"), b)
	fc := codegraph.NewFileNode(codegraph.NodeID("file:c.go"), c)
	x := codegraph.NewFunctionNode(codegraph.NodeID("fn:a.go:X:1"), a, "X", 1, 2)
	y := codegraph.NewFunctionNode(codegraph.NodeID("fn:c.go:Y:1"), c, "Y", 1, 2)
	z := codegraph.NewFunctionNode(codegraph.NodeID("fn:b.go:Z:1"), b, "Z", 1, 2)

	for _, err := range []error{
		store.AddNode(fa),
		store.AddNode(fb),
		store.AddNode(fc),
		store.AddNode(x),
		store.AddNode(y),
		store.AddNode(z),
		store.AddEdge(codegraph.NewEdge(fa.ID, x.ID, codegraph.EdgeDefines)),
		store.AddEdge(codegraph.NewEdge(fc.ID, y.ID, codegraph.EdgeDefines)),
		store.AddEdge(codegraph.NewEdge(fb.ID, z.ID, codegraph.EdgeDefines)),
		store.AddEdge(codegraph.NewEdge(x.ID, y.ID, codegraph.EdgeCalls)),
		store.AddEdge(codegraph.NewEdge(y.ID, z.ID, codegraph.EdgeCalls)),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	g, err := store.Build()
	if err != nil {
		t.Fatal(err)
	}

	hops, ok := codegraph.MinCallHopsAcrossFiles(g, []domain.RepoRelPath{a}, []domain.RepoRelPath{b})
	if !ok {
		t.Fatal("expected hops across files")
	}
	if hops != 2 {
		t.Fatalf("expected 2 hops X→Y→Z, got %d", hops)
	}
}
