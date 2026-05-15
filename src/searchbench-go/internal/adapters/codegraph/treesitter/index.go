//go:build cgo

package treesitter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"

	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// BuildDirectory walks absRepoRoot for *.go sources (skipping dot dirs, vendor,
// and .git) and returns an in-memory graph with file nodes, function nodes,
// defines edges, and same-directory call edges resolved by callee identifier name.
func BuildDirectory(absRepoRoot string) (*codegraph.Store, error) {
	rootInfo, err := os.Stat(absRepoRoot)
	if err != nil {
		return nil, fmt.Errorf("treesitter index root: %w", err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("treesitter index root %q is not a directory", absRepoRoot)
	}

	lang := golang.GetLanguage()
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	defer parser.Close()

	type fileJob struct {
		abs  string
		rel  domain.RepoRelPath
		data []byte
	}

	var jobs []fileJob
	err = filepath.WalkDir(absRepoRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			base := d.Name()
			if strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			if base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".go") {
			return nil
		}
		relPath, err := filepath.Rel(absRepoRoot, path)
		if err != nil {
			return err
		}
		repoRel := domain.CanonicalizePath(filepath.ToSlash(relPath))
		if repoRel == "" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		jobs = append(jobs, fileJob{abs: path, rel: repoRel, data: b})
		return nil
	})
	if err != nil {
		return nil, err
	}

	dirKey := func(absFile string) string { return filepath.Clean(filepath.Dir(absFile)) }

	type dirSyms struct {
		funcs map[string]codegraph.NodeID
	}
	dirFuncs := make(map[string]*dirSyms)

	for _, job := range jobs {
		dk := dirKey(job.abs)
		if dirFuncs[dk] == nil {
			dirFuncs[dk] = &dirSyms{funcs: make(map[string]codegraph.NodeID)}
		}
		tree, err := parser.ParseCtx(context.Background(), nil, job.data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", job.rel, err)
		}
		root := tree.RootNode()
		if root.HasError() {
			continue
		}
		collectFuncSymbols(*root, job.data, job.rel, dirFuncs[dk].funcs)
	}

	store := codegraph.NewStore()
	addedFiles := make(map[domain.RepoRelPath]struct{})

	for _, job := range jobs {
		tree, err := parser.ParseCtx(context.Background(), nil, job.data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", job.rel, err)
		}
		root := tree.RootNode()
		if root.HasError() {
			continue
		}
		sym := dirFuncs[dirKey(job.abs)].funcs
		if err := materializeFile(store, addedFiles, *root, job.data, job.rel, sym); err != nil {
			return nil, err
		}
	}

	return store, nil
}

func collectFuncSymbols(n sitter.Node, src []byte, file domain.RepoRelPath, out map[string]codegraph.NodeID) {
	switch n.Type() {
	case "function_declaration", "method_declaration":
		id := declFuncID(n, src, file)
		if id == "" {
			break
		}
		name := declName(n, src)
		if name != "" {
			out[name] = id
		}
	}
	for i := 0; i < int(n.NamedChildCount()); i++ {
		ch := n.NamedChild(i)
		if ch != nil {
			collectFuncSymbols(*ch, src, file, out)
		}
	}
}

func materializeFile(store *codegraph.Store, addedFiles map[domain.RepoRelPath]struct{}, root sitter.Node, src []byte, file domain.RepoRelPath, sym map[string]codegraph.NodeID) error {
	fileID := codegraph.NodeID("file:" + string(file))
	if _, ok := addedFiles[file]; !ok {
		if err := store.AddNode(codegraph.NewFileNode(fileID, file)); err != nil {
			return err
		}
		addedFiles[file] = struct{}{}
	}

	var walkErr error
	var walk func(sitter.Node)
	walk = func(n sitter.Node) {
		if walkErr != nil {
			return
		}
		switch n.Type() {
		case "function_declaration", "method_declaration":
			id := declFuncID(n, src, file)
			if id == "" {
				break
			}
			fn := buildFuncNode(n, src, file, id)
			if err := store.AddNode(fn); err != nil {
				walkErr = err
				return
			}
			if err := store.AddEdge(codegraph.NewEdge(fileID, id, codegraph.EdgeDefines)); err != nil {
				walkErr = err
				return
			}
			if body := n.ChildByFieldName("body"); body != nil {
				scanCalls(store, *body, src, id, sym)
			}
			return
		}
		for i := 0; i < int(n.NamedChildCount()); i++ {
			ch := n.NamedChild(i)
			if ch != nil {
				walk(*ch)
			}
		}
	}
	walk(root)
	return walkErr
}

func scanCalls(store *codegraph.Store, n sitter.Node, src []byte, caller codegraph.NodeID, sym map[string]codegraph.NodeID) {
	if n.Type() == "call_expression" {
		if fun := n.ChildByFieldName("function"); fun != nil && fun.Type() == "identifier" {
			name := strings.TrimSpace(fun.Content(src))
			if callee, ok := sym[name]; ok && callee != caller {
				_ = store.AddEdge(codegraph.NewEdge(caller, callee, codegraph.EdgeCalls))
			}
		}
	}
	for i := 0; i < int(n.NamedChildCount()); i++ {
		ch := n.NamedChild(i)
		if ch != nil {
			scanCalls(store, *ch, src, caller, sym)
		}
	}
}

func declName(n sitter.Node, src []byte) string {
	name := n.ChildByFieldName("name")
	if name == nil {
		return ""
	}
	return strings.TrimSpace(name.Content(src))
}

func declFuncID(n sitter.Node, src []byte, file domain.RepoRelPath) codegraph.NodeID {
	name := n.ChildByFieldName("name")
	if name == nil {
		return ""
	}
	nm := strings.TrimSpace(name.Content(src))
	if nm == "" {
		return ""
	}
	line := int(name.StartPoint().Row) + 1
	return codegraph.NodeID(fmt.Sprintf("fn:%s:%s:%d", file, nm, line))
}

func buildFuncNode(decl sitter.Node, src []byte, file domain.RepoRelPath, id codegraph.NodeID) codegraph.Node {
	name := declName(decl, src)
	start := 1
	end := 1
	if nn := decl.ChildByFieldName("name"); nn != nil {
		start = int(nn.StartPoint().Row) + 1
		end = start
	}
	if body := decl.ChildByFieldName("body"); body != nil {
		end = int(body.EndPoint().Row) + 1
	}
	return codegraph.NewFunctionNode(id, file, name, start, end)
}
