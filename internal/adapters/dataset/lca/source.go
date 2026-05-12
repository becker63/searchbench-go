package lca

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	gitmaterialize "github.com/becker63/searchbench-go/internal/adapters/materialize/git"
	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

type Source struct{}

// NewSource returns a MatchSource that reads JetBrains LCA JSONL fixtures.
func NewSource() Source { return Source{} }

// Matches loads manifest-local JSONL, normalizes rows, sorts by MatchID, and
// applies MaxItems windowing when set.
func (Source) Matches(ctx context.Context, req dataset.Request) (domain.NonEmpty[domain.MatchSpec], error) {
	if err := ctx.Err(); err != nil {
		return domain.NonEmpty[domain.MatchSpec]{}, err
	}
	if !IsJetBrainsDataset(req) {
		return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf(
			"lca adapter: dataset kind %q name %q is not JetBrains LCA",
			req.Kind, req.Name,
		)
	}

	path, err := ResolveJetBrainsLCASlicePath(req.ManifestDir, req.Name, req.Config, req.Split)
	if err != nil {
		return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: %w", err)
	}
	rawRows, err := loadRowsJSONL(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: dataset file missing: %s", path)
		}
		return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: load %s: %w", path, err)
	}

	datasetName := strings.TrimSpace(req.Name)
	if datasetName == "" {
		datasetName = JetBrainsLCADatasetName
	}
	config := strings.TrimSpace(req.Config)
	split := strings.TrimSpace(req.Split)

	var tasks []domain.LCATask
	for _, raw := range rawRows {
		if err := ctx.Err(); err != nil {
			return domain.NonEmpty[domain.MatchSpec]{}, err
		}
		var row domain.LCAHFRow
		if err := json.Unmarshal(raw, &row); err != nil {
			return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: parse row: %w", err)
		}
		task, err := row.ToTask(datasetName, config, split)
		if err != nil {
			return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: normalize task: %w", err)
		}
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return string(tasks[i].MatchID()) < string(tasks[j].MatchID())
	})

	var mz *gitmaterialize.Materializer
	if string(req.MaterializeCacheDir) != "" {
		mz = &gitmaterialize.Materializer{}
	}

	var specs []domain.MatchSpec
	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			return domain.NonEmpty[domain.MatchSpec]{}, err
		}
		t := task
		if mz != nil {
			mreq := gitmaterialize.MaterializeRequest{
				Task:      t,
				CacheDir:  req.MaterializeCacheDir,
				RemoteURL: strings.TrimSpace(req.MaterializeRemoteURL),
			}
			res, err := mz.Materialize(ctx, mreq)
			if err != nil {
				return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: materialize %s: %w", t.MatchID(), err)
			}
			t = t.WithRepo(res.RootPath)
		}
		specs = append(specs, t.MatchSpec())
	}

	if req.MaxItems != nil {
		if *req.MaxItems <= 0 {
			return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: maxItems must be positive")
		}
		if len(specs) > *req.MaxItems {
			specs = specs[:*req.MaxItems]
		}
	}

	out, err := domain.NonEmptyFromSlice(specs)
	if err != nil {
		return domain.NonEmpty[domain.MatchSpec]{}, fmt.Errorf("lca adapter: %w", err)
	}
	return out, nil
}
