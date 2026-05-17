package usage

import "testing"

func TestHashRegistryRecordsDistinctHashes(t *testing.T) {
	t.Parallel()
	var r HashRegistry
	r.RecordRequest([]byte("req-a"))
	r.RecordResponse([]byte("resp-b"))
	req, resp := r.Snapshot()
	if len(req) != 1 || len(resp) != 1 || req[0] == resp[0] {
		t.Fatalf("req=%v resp=%v", req, resp)
	}
}
