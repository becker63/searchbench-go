package codegraph

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// MinCallHopsAcrossFiles reports the smallest shortest-path hop count between any
// function node anchored in fromFiles and any function node anchored in toFiles,
// considering only directed EdgeCalls edges.
//
// Files are matched via Node.RepoFile; only NodeFunction kinds participate.
// If either side has no function nodes in those files, it returns ok=false.
func MinCallHopsAcrossFiles(g Graph, fromFiles, toFiles []domain.RepoRelPath) (hops int, ok bool) {
	filter := EdgeFilter{Kinds: []EdgeKind{EdgeCalls}}

	fromIDs := functionNodesInFiles(g, fromFiles)
	toIDs := functionNodesInFiles(g, toFiles)
	if len(fromIDs) == 0 || len(toIDs) == 0 {
		return 0, false
	}

	best := 0
	found := false
	for _, a := range fromIDs {
		for _, b := range toIDs {
			if a == b {
				continue
			}
			p, okp := g.ShortestPath(a, b, filter)
			if !okp || len(p.Nodes) == 0 {
				continue
			}
			h := p.Hops()
			if !found || h < best {
				best = h
				found = true
			}
		}
	}
	return best, found
}

func functionNodesInFiles(g Graph, files []domain.RepoRelPath) []NodeID {
	seen := make(map[domain.RepoRelPath]struct{}, len(files))
	for _, f := range files {
		seen[f] = struct{}{}
	}

	var out []NodeID
	for file := range seen {
		for _, id := range g.NodesByFile(file) {
			n, ok := g.Node(id)
			if !ok || n.Kind != NodeFunction {
				continue
			}
			out = append(out, id)
		}
	}
	return out
}
