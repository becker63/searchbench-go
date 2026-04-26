package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestLCATaskIdentityTaskIDDeterministic(t *testing.T) {
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

	got := identity.TaskID()
	want := TaskID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1")
	if got != want {
		t.Fatalf("TaskID() = %q, want %q", got, want)
	}
	if identity.TaskID() != got {
		t.Fatalf("TaskID() is not deterministic: got %q then %q", got, identity.TaskID())
	}
}

func TestLCATaskIdentityIssueKeyFallbacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		identity LCATaskIdentity
		want     TaskID
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
			want: TaskID("d:c:s:square/okhttp@abc123:https://example.test/issues/1"),
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
			want: TaskID("d:c:s:square/okhttp@abc123:https://example.test/pulls/2"),
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
			want: TaskID("d:c:s:square/okhttp@abc123:unknown-issue"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.identity.TaskID(); got != tt.want {
				t.Fatalf("TaskID() = %q, want %q", got, tt.want)
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
			wantID := TaskID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1")
			if row.IssueURL == "" {
				wantID = TaskID("searchbench/lca:python:dev:square/okhttp@abc123:unknown-issue")
			}
			if got := task.TaskID(); got != wantID {
				t.Fatalf("task.TaskID() = %q, want %q", got, wantID)
			}
			if got := task.ChangedFiles(); len(got) != 1 || got[0] != "src/main.kt" {
				t.Fatalf("ChangedFiles() = %#v, want [\"src/main.kt\"]", got)
			}
		})
	}
}
