package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestLCATaskIdentityMatchIDDeterministic(t *testing.T) {
	t.Parallel()

	identity := LCATaskIdentity{
		DatasetName:   "searchbench/lca",
		DatasetConfig: "python",
		DatasetSplit:  "dev",
		RepoOwner:     "Square",
		RepoName:      "OkHttp",
		BaseSHA:       "ABC123",
		IssueURL:      "https://example.test/issues/1",
	}

	got := identity.MatchID()
	want := MatchID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1")
	if got != want {
		t.Fatalf("MatchID() = %q, want %q", got, want)
	}
	if identity.MatchID() != got {
		t.Fatalf("MatchID() is not deterministic: got %q then %q", got, identity.MatchID())
	}
}

func TestLCATaskIdentityIssueKeyFallbacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		identity LCATaskIdentity
		want     MatchID
	}{
		{
			name: "issue url preferred",
			identity: LCATaskIdentity{
				DatasetName:   "d",
				DatasetConfig: "c",
				DatasetSplit:  "s",
				RepoOwner:     "square",
				RepoName:      "okhttp",
				BaseSHA:       "abc123",
				IssueURL:      "https://example.test/issues/1",
				PullURL:       "https://example.test/pulls/2",
			},
			want: MatchID("d:c:s:square/okhttp@abc123:https://example.test/issues/1"),
		},
		{
			name: "pull url fallback",
			identity: LCATaskIdentity{
				DatasetName:   "d",
				DatasetConfig: "c",
				DatasetSplit:  "s",
				RepoOwner:     "square",
				RepoName:      "okhttp",
				BaseSHA:       "abc123",
				PullURL:       "https://example.test/pulls/2",
			},
			want: MatchID("d:c:s:square/okhttp@abc123:https://example.test/pulls/2"),
		},
		{
			name: "unknown issue fallback",
			identity: LCATaskIdentity{
				DatasetName:   "d",
				DatasetConfig: "c",
				DatasetSplit:  "s",
				RepoOwner:     "square",
				RepoName:      "okhttp",
				BaseSHA:       "abc123",
			},
			want: MatchID("d:c:s:square/okhttp@abc123:unknown-issue"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.identity.MatchID(); got != tt.want {
				t.Fatalf("MatchID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCanonicalizePaths(t *testing.T) {
	t.Parallel()

	got := CanonicalizePaths([]string{
		"src/Main.kt",
		"./src/Main.kt",
		"/src/Main.kt",
		`src\Main.kt`,
	})
	want := []RepoRelPath{"src/main.kt"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLCAHFRowToTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "changed files array",
			payload: `{
				"repo_owner":"square",
				"repo_name":"okhttp",
				"base_sha":"abc123",
				"issue_title":"t",
				"issue_body":"b",
				"issue_url":"https://example.test/issues/1",
				"changed_files":["./src/Main.kt","src\\Main.kt"]
			}`,
		},
		{
			name: "changed files stringified list",
			payload: `{
				"repo_owner":"square",
				"repo_name":"okhttp",
				"base_sha":"abc123",
				"issue_title":"t",
				"issue_body":"b",
				"changed_files":"['src/Main.kt']"
			}`,
		},
		{
			name: "missing base sha",
			payload: `{
				"repo_owner":"square",
				"repo_name":"okhttp",
				"issue_title":"t",
				"issue_body":"b",
				"changed_files":["src/Main.kt"]
			}`,
			wantErr: "base_sha",
		},
		{
			name: "missing changed files",
			payload: `{
				"repo_owner":"square",
				"repo_name":"okhttp",
				"base_sha":"abc123",
				"issue_title":"t",
				"issue_body":"b"
			}`,
			wantErr: "changed_files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var row LCAHFRow
			err := json.Unmarshal([]byte(tt.payload), &row)
			if err == nil {
				_, err = row.ToTask("searchbench/lca", "python", "dev")
			}

			if tt.wantErr != "" {
				var schemaErr *SchemaError
				if !errors.As(err, &schemaErr) {
					t.Fatalf("expected SchemaError, got %v", err)
				}
				if schemaErr.Field != tt.wantErr {
					t.Fatalf("SchemaError.Field = %q, want %q", schemaErr.Field, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			task, err := row.ToTask("searchbench/lca", "python", "dev")
			if err != nil {
				t.Fatalf("unexpected conversion error: %v", err)
			}
			wantID := MatchID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1")
			if row.IssueURL == "" {
				wantID = MatchID("searchbench/lca:python:dev:square/okhttp@abc123:unknown-issue")
			}
			if got := task.MatchID(); got != wantID {
				t.Fatalf("task.MatchID() = %q, want %q", got, wantID)
			}
			if got := task.ChangedFiles(); len(got) != 1 || got[0] != "src/main.kt" {
				t.Fatalf("ChangedFiles() = %#v, want [\"src/main.kt\"]", got)
			}
		})
	}
}
