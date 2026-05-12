package gitmaterialize

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// Materializer prepares local checkouts for LCA tasks.
type Materializer struct {
	Git GitRunner
}

// Materialize clones or reuses a bare mirror and checks out ref (base SHA) into
// a deterministic per-(owner,name,sha) working tree under [MaterializeRequest.CacheDir].
func (m *Materializer) Materialize(ctx context.Context, req MaterializeRequest) (MaterializedRepo, error) {
	if err := req.Task.Validate(); err != nil {
		return MaterializedRepo{}, newErr(FailureInvalidTaskRepo, err)
	}

	remote := strings.TrimSpace(req.RemoteURL)
	if remote == "" {
		remote = DefaultGitHubRemoteURL(req.Task.Identity)
	}
	if err := validateRemoteURL(remote); err != nil {
		return MaterializedRepo{}, newErr(FailureInvalidRepoURL, err)
	}

	runner := m.Git
	if runner == nil {
		runner = ExecGitRunner{}
	}

	cacheRoot := filepath.Clean(string(req.CacheDir))
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return MaterializedRepo{}, newErr(FailureCachePermission, err)
	}

	id := req.Task.Identity
	key := CacheKeyFromTask(req.Task)
	checkoutPath := filepath.Join(cacheRoot, "checkouts", key)
	mirrorPath := filepath.Join(cacheRoot, "mirrors", mirrorDirName(id.RepoOwner, id.RepoName))

	if req.ForceRefresh {
		if err := removePathIfExists(checkoutPath); err != nil {
			return MaterializedRepo{}, newErr(FailureFilesystemError, err)
		}
	}

	if err := ensureMirror(ctx, runner, remote, mirrorPath); err != nil {
		return MaterializedRepo{}, err
	}

	if err := ensureSHAInMirror(ctx, runner, mirrorPath, remote, id.BaseSHA); err != nil {
		return MaterializedRepo{}, err
	}

	hit, err := cacheHit(ctx, runner, checkoutPath, id.BaseSHA)
	if err != nil {
		return MaterializedRepo{}, err
	}
	if hit {
		return buildResult(req.Task, remote, key, checkoutPath), nil
	}

	if err := removePathIfExists(checkoutPath); err != nil {
		return MaterializedRepo{}, newErr(FailureFilesystemError, err)
	}

	_, stderr, code, execErr := runner.RunGit(ctx, "", "clone", "--shared", mirrorPath, checkoutPath)
	if execErr != nil {
		return MaterializedRepo{}, newErr(FailureFilesystemError, execErr)
	}
	if code != 0 {
		return MaterializedRepo{}, newErr(FailureCloneFailed, fmt.Errorf("git clone: %s", strings.TrimSpace(stderr)))
	}

	if err := checkoutAtSHA(ctx, runner, checkoutPath, id.BaseSHA); err != nil {
		return MaterializedRepo{}, err
	}

	if err := markWorktreeReadOnly(checkoutPath); err != nil {
		return MaterializedRepo{}, newErr(FailureFilesystemError, err)
	}

	return buildResult(req.Task, remote, key, checkoutPath), nil
}

func buildResult(task domain.LCATask, remote, key, checkoutPath string) MaterializedRepo {
	id := task.Identity
	return MaterializedRepo{
		TaskID:         task.MatchID(),
		RepoOwner:      strings.TrimSpace(id.RepoOwner),
		RepoName:       strings.TrimSpace(id.RepoName),
		BaseSHA:        strings.TrimSpace(id.BaseSHA),
		RepoURL:        remote,
		RootPath:       domain.HostPath(checkoutPath),
		CacheKey:       key,
		MaterializedAt: time.Now().UTC(),
	}
}

func mirrorDirName(owner, name string) string {
	return pathSegment(owner) + "_" + pathSegment(name) + ".git"
}

func pathSegment(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "_"
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func validateRemoteURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "http", "https", "file":
		return nil
	default:
		return fmt.Errorf("unsupported remote scheme %q (want http(s) or file)", u.Scheme)
	}
}

func ensureMirror(ctx context.Context, runner GitRunner, remote, mirrorPath string) error {
	fi, statErr := os.Stat(mirrorPath)
	if statErr == nil && fi.IsDir() {
		return nil
	}

	parent := filepath.Dir(mirrorPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return newErr(FailureCachePermission, err)
	}

	_, stderr, code, execErr := runner.RunGit(ctx, "", "clone", "--mirror", remote, mirrorPath)
	if execErr != nil {
		return newErr(FailureFilesystemError, execErr)
	}
	if code != 0 {
		return newErr(FailureCloneFailed, fmt.Errorf("git clone --mirror: %s", strings.TrimSpace(stderr)))
	}
	return nil
}

func ensureSHAInMirror(ctx context.Context, runner GitRunner, mirrorPath string, remote, sha string) error {
	ref := strings.TrimSpace(sha)
	if ref == "" {
		return newErr(FailureInvalidTaskRepo, fmt.Errorf("base_sha is empty"))
	}

	if shaExistsBare(ctx, runner, mirrorPath, ref) {
		return nil
	}

	_, stderr, code, execErr := runner.RunGit(ctx, "",
		"--git-dir", mirrorPath,
		"fetch", "--depth", "1", remote, ref,
	)
	if execErr != nil {
		return newErr(FailureFilesystemError, execErr)
	}

	if code != 0 {
		lower := strings.ToLower(stderr)
		if strings.Contains(lower, "could not resolve") ||
			strings.Contains(lower, "invalid") ||
			strings.Contains(lower, "fatal: couldn't find remote ref") {
			return newErr(FailureMissingBaseSHA, fmt.Errorf("missing object for ref %s: %s", ref, strings.TrimSpace(stderr)))
		}
		return newErr(FailureFetchFailed, fmt.Errorf("git fetch %s: %s", ref, strings.TrimSpace(stderr)))
	}

	if shaExistsBare(ctx, runner, mirrorPath, ref) {
		return nil
	}

	_, stderrDeep, codeDeep, execErrDeep := runner.RunGit(ctx, "",
		"--git-dir", mirrorPath,
		"fetch", remote, ref,
	)
	if execErrDeep != nil {
		return newErr(FailureFilesystemError, execErrDeep)
	}
	if codeDeep != 0 || !shaExistsBare(ctx, runner, mirrorPath, ref) {
		return newErr(FailureMissingBaseSHA, fmt.Errorf("object not found after fetch for ref %s: %s",
			ref, strings.TrimSpace(stderrDeep)))
	}
	return nil
}

func shaExistsBare(ctx context.Context, runner GitRunner, bareDir string, sha string) bool {
	_, _, code, execErr := runner.RunGit(ctx, "",
		"--git-dir", bareDir,
		"rev-parse", "-q", "--verify", sha+"^{commit}",
	)
	return execErr == nil && code == 0
}

func cacheHit(ctx context.Context, runner GitRunner, checkoutPath string, wantSHA string) (bool, error) {
	fi, statErr := os.Stat(filepath.Join(checkoutPath, ".git"))
	if statErr != nil || !fi.IsDir() {
		return false, nil
	}

	head, _, headCode, headErr := runner.RunGit(ctx, checkoutPath, "rev-parse", "HEAD")
	if headErr != nil {
		return false, newErr(FailureFilesystemError, headErr)
	}
	if headCode != 0 {
		return false, nil
	}

	target, _, wantCode, wantErr := runner.RunGit(ctx, checkoutPath, "rev-parse", "-q", "--verify", wantSHA+"^{commit}")
	if wantErr != nil {
		return false, newErr(FailureFilesystemError, wantErr)
	}
	if wantCode != 0 {
		return false, nil
	}

	return strings.TrimSpace(head) == strings.TrimSpace(target), nil
}

func checkoutAtSHA(ctx context.Context, runner GitRunner, checkoutPath string, sha string) error {
	_, stderr, code, execErr := runner.RunGit(ctx, checkoutPath, "checkout", "-q", "--detach", sha)
	if execErr != nil {
		return newErr(FailureFilesystemError, execErr)
	}
	if code != 0 {
		lower := strings.ToLower(stderr)
		if strings.Contains(lower, "unknown revision") ||
			strings.Contains(lower, "did not match any file") {
			return newErr(FailureMissingBaseSHA, fmt.Errorf("%s", strings.TrimSpace(stderr)))
		}
		return newErr(FailureCheckoutFailed, fmt.Errorf("git checkout: %s", strings.TrimSpace(stderr)))
	}
	return nil
}

func markWorktreeReadOnly(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return os.Chmod(path, 0o755)
		}
		return os.Chmod(path, 0o444)
	})
}

func removePathIfExists(p string) error {
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return os.RemoveAll(p)
}
