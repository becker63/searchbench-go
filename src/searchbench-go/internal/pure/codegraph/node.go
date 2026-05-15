package codegraph

import "github.com/becker63/searchbench-go/internal/pure/domain"

// NodeID identifies a node in the Searchbench evaluation graph.
//
// It is intentionally local to codegraph instead of reusing domain.NodeID so
// graph concepts do not leak into the broader domain package too early.
type NodeID string

func (id NodeID) String() string {
	return string(id)
}

// NodeKind identifies the semantic kind of a code graph node.
type NodeKind string

const (
	NodeFile     NodeKind = "file"
	NodeFunction NodeKind = "function"
	NodeSymbol   NodeKind = "symbol"
	NodeImport   NodeKind = "import"
)

// FileNode represents a repo-relative source file.
type FileNode struct {
	Path domain.RepoRelPath `json:"path"`
}

// FunctionNode represents a function/method definition.
type FunctionNode struct {
	File      domain.RepoRelPath `json:"file"`
	Name      string             `json:"name"`
	StartLine int                `json:"start_line"`
	EndLine   int                `json:"end_line"`
}

// SymbolNode represents a named symbol that is not necessarily a function.
type SymbolNode struct {
	File domain.RepoRelPath `json:"file"`
	Name string             `json:"name"`
}

// ImportNode represents an import/module dependency observed in a file.
type ImportNode struct {
	File   domain.RepoRelPath `json:"file"`
	Module string             `json:"module"`
}

// Node is a tagged-union-ish representation of code graph nodes.
//
// Exactly one payload should be non-nil and should match Kind. Constructors
// below enforce that convention for normal code paths.
type Node struct {
	ID       NodeID        `json:"id"`
	Kind     NodeKind      `json:"kind"`
	File     *FileNode     `json:"file,omitempty"`
	Function *FunctionNode `json:"function,omitempty"`
	Symbol   *SymbolNode   `json:"symbol,omitempty"`
	Import   *ImportNode   `json:"import,omitempty"`
}

// NewFileNode constructs a file node.
func NewFileNode(id NodeID, path domain.RepoRelPath) Node {
	return Node{
		ID:   id,
		Kind: NodeFile,
		File: &FileNode{Path: path},
	}
}

// NewFunctionNode constructs a function node.
func NewFunctionNode(id NodeID, file domain.RepoRelPath, name string, startLine, endLine int) Node {
	return Node{
		ID:   id,
		Kind: NodeFunction,
		Function: &FunctionNode{
			File:      file,
			Name:      name,
			StartLine: startLine,
			EndLine:   endLine,
		},
	}
}

// NewSymbolNode constructs a symbol node.
func NewSymbolNode(id NodeID, file domain.RepoRelPath, name string) Node {
	return Node{
		ID:   id,
		Kind: NodeSymbol,
		Symbol: &SymbolNode{
			File: file,
			Name: name,
		},
	}
}

// NewImportNode constructs an import node.
func NewImportNode(id NodeID, file domain.RepoRelPath, module string) Node {
	return Node{
		ID:   id,
		Kind: NodeImport,
		Import: &ImportNode{
			File:   file,
			Module: module,
		},
	}
}

// RepoFile returns the file associated with the node when one exists.
func (n Node) RepoFile() (domain.RepoRelPath, bool) {
	switch n.Kind {
	case NodeFile:
		if n.File == nil {
			return "", false
		}
		return n.File.Path, true
	case NodeFunction:
		if n.Function == nil {
			return "", false
		}
		return n.Function.File, true
	case NodeSymbol:
		if n.Symbol == nil {
			return "", false
		}
		return n.Symbol.File, true
	case NodeImport:
		if n.Import == nil {
			return "", false
		}
		return n.Import.File, true
	default:
		return "", false
	}
}
