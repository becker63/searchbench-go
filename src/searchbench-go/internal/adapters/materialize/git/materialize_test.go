package gitmaterialize

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestMaterialize_LocalRepo_CheckoutSHA_AndReuse(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	src := filepath.Join(bundle, "src")
	mirror := filepath.Join(bundle, "upstream.git")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, src, "init")
	runGit(t, src, "config", "user.email", "fixture@test")
	runGit(t, src, "config", "user.name", "fixture")
	writeFile(t, filepath.Join(src, "README.md"), "# hi\n")

	runGit(t, src, "add", ".")
	runGit(t, src, "commit", "-m", "c1")

	sha1 := revParseHEAD(t, src)

	writeFile(t, filepath.Join(src, "two.txt"), "second\n")

	runGit(t, src, "add", ".")
	runGit(t, src, "commit", "-m", "c2")
	sha2 := revParseHEAD(t, src)

	runGit(t, bundle, "-C", src, "remote", "add", "origin", mirror)

	runGit(t, bundle, "clone", "--mirror", src, mirror)

	cache := filepath.Join(bundle, "cache")
	task := fixtureTask(t, "Acme", "Demo", sha1)

	mz := Materializer{}

	res1, err := mz.Materialize(context.Background(),
		MaterializeRequest{Task: task, CacheDir: domain.HostPath(cache), RemoteURL: fileURL(t, mirror), ForceRefresh: false})
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}

	assertFileReadable(t, filepath.Join(string(res1.RootPath), "README.md"))
	readmeMissing := filepath.Join(string(res1.RootPath), "two.txt")
	if _, statErr := os.Stat(readmeMissing); statErr == nil {
		t.Fatalf("unexpected second commit file at earlier sha checkout")
	} else if !os.IsNotExist(statErr) {
		t.Fatal(statErr)
	}

	res1b, err := mz.Materialize(context.Background(),
		MaterializeRequest{Task: task, CacheDir: domain.HostPath(cache), RemoteURL: fileURL(t, mirror), ForceRefresh: false})
	if err != nil {
		t.Fatalf("second materialize: %v", err)
	}
	if res1.RootPath != res1b.RootPath {
		t.Fatalf("cache reuse: paths differ")
	}

	task2 := fixtureTask(t, "acme", "demo", sha2)

	res2, err := mz.Materialize(context.Background(), MaterializeRequest{
		Task:         task2,
		CacheDir:     domain.HostPath(cache),
		RemoteURL:    fileURL(t, mirror),
		ForceRefresh: false,
	})
	if err != nil {
		t.Fatalf("materialize sha2: %v", err)
	}
	if _, err := os.Stat(filepath.Join(string(res2.RootPath), "two.txt")); err != nil {
		t.Fatalf("expected second-commit file at sha2: %v", err)
	}
	if strings.EqualFold(string(res1.RootPath), string(res2.RootPath)) {
		t.Fatalf("expected distinct checkout dirs for distinct shas")
	}
}

func TestMaterialize_InvalidSHA_CheckoutFails(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	src := filepath.Join(bundle, "src")
	mirror := filepath.Join(bundle, "upstream.git")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, src, "init")
	runGit(t, src, "config", "user.email", "fixture@test")
	runGit(t, src, "config", "user.name", "fixture")
	writeFile(t, filepath.Join(src, "README.md"), "# hi\n")

	runGit(t, src, "add", ".")
	runGit(t, src, "commit", "-m", "c1")
	runGit(t, bundle, "clone", "--mirror", src, mirror)

	task := fixtureTask(t, "acme", "demo", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	mz := Materializer{}
	_, err := mz.Materialize(context.Background(), MaterializeRequest{
		Task:         task,
		CacheDir:     domain.HostPath(filepath.Join(bundle, "cache")),
		RemoteURL:    fileURL(t, mirror),
		ForceRefresh: false,
	})
	if err == nil {
		t.Fatal("expected error for missing sha")
	}
	kind, ok := AsFailureKind(err)
	if !ok {
		t.Fatalf("expected typed error, got %v", err)
	}
	if kind != FailureFetchFailed && kind != FailureMissingBaseSHA {
		t.Fatalf("unexpected kind %s for %v", kind, err)
	}
}

func TestMaterialize_InvalidTask(t *testing.T) {
	t.Parallel()

	task := domain.LCATask{}
	mz := Materializer{}
	_, err := mz.Materialize(context.Background(), NewMaterializeRequest(task, domain.HostPath(t.TempDir()), false))
	if err == nil {
		t.Fatal("expected validation error")
	}
	kind, ok := AsFailureKind(err)
	if !ok || kind != FailureInvalidTaskRepo {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestMaterialize_InvalidRemoteURL(t *testing.T) {
	t.Parallel()

	task := fixtureTask(t, "a", "b", revParseHEAD(t, bootstrapOneCommitRepo(t)))
	mz := Materializer{}

	_, err := mz.Materialize(context.Background(), MaterializeRequest{
		Task:         task,
		CacheDir:     domain.HostPath(t.TempDir()),
		RemoteURL:    "gitlab://oops/wrong.git",
		ForceRefresh: false,
	})
	if err == nil {
		t.Fatal("expected invalid url error")
	}
	kind, ok := AsFailureKind(err)
	if !ok || kind != FailureInvalidRepoURL {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestCacheKey_FromTask(t *testing.T) {
	t.Parallel()

	id := domain.LCATaskIdentity{
		DatasetName: "dn", DatasetConfig: "dc", DatasetSplit: "ds",
		RepoOwner: "Foo", RepoName: "Bar", BaseSHA: "AbCdEf01",
		IssueURL: "https://issue/1",
	}
	ctx := domain.LCAContext{IssueTitle: "t", IssueBody: "b"}
	gold, err := domain.NewLCAGold([]string{"a.go"})
	if err != nil {
		t.Fatal(err)
	}
	task := domain.LCATask{Identity: id, Context: ctx, Gold: gold}

	got := CacheKeyFromTask(task)
	want := filepath.Join("foo", "bar", "abcdef01")
	if got != want {
		t.Fatalf("CacheKeyFromTask got %q want %q", got, want)
	}
}

func TestMaterialize_ResultFields(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	src := filepath.Join(bundle, "src")
	mirror := filepath.Join(bundle, "upstream.git")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, src, "init")
	runGit(t, src, "config", "user.email", "fixture@test")
	runGit(t, src, "config", "user.name", "fixture")
	writeFile(t, filepath.Join(src, "x.txt"), "x\n")
	runGit(t, src, "add", ".")
	runGit(t, src, "commit", "-m", "c1")
	sha := revParseHEAD(t, src)
	runGit(t, bundle, "clone", "--mirror", src, mirror)

	task := fixtureTask(t, "Acme", "Demo", sha)
	mz := Materializer{}
	res, err := mz.Materialize(context.Background(), MaterializeRequest{
		Task:         task,
		CacheDir:     domain.HostPath(filepath.Join(bundle, "cache")),
		RemoteURL:    fileURL(t, mirror),
		ForceRefresh: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TaskID != task.MatchID() {
		t.Fatalf("task id: got %q want %q", res.TaskID, task.MatchID())
	}
	if res.RepoOwner != "Acme" || res.RepoName != "Demo" || res.BaseSHA != sha {
		t.Fatalf("repo fields mismatch: %#v vs want owner Acme name Demo sha %s", res, sha)
	}
	st, statErr := os.Stat(string(res.RootPath))
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !st.IsDir() {
		t.Fatalf("RootPath must be directory")
	}
	if res.MaterializedAt.IsZero() {
		t.Fatal("materialized timestamp missing")
	}
	if filepath.Base(string(res.RootPath)) == "" {
		t.Fatal("empty basename")
	}
}

func bootstrapOneCommitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "fixture@test")
	runGit(t, dir, "config", "user.name", "fixture")
	writeFile(t, filepath.Join(dir, "a.txt"), "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "x")
	return dir
}

func fixtureTask(t *testing.T, owner, name, sha string) domain.LCATask {
	t.Helper()

	task, taskErr := (domain.LCAHFRow{
		RepoOwner:    owner,
		RepoName:     name,
		BaseSHA:      sha,
		IssueTitle:   "title",
		IssueBody:    "body",
		ChangedFiles: []string{"a.go"},
		IssueURL:     "https://example.com/issues/1",
	}).ToTask("jet", "py", "train")
	if taskErr != nil {
		t.Fatal(taskErr)
	}
	return task
}

func assertFileReadable(t *testing.T, path string) {
	t.Helper()
	fi, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if fi.Mode().IsRegular() && fi.Mode().Perm()&0400 == 0 {
		t.Fatalf("%s missing user-read bit: %v", path, fi.Mode())
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func revParseHEAD(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fileURL(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return u.String()
}
