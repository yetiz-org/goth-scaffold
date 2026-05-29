package internal

import (
	"bytes"
	"encoding/json"
	"testing"
)

// FlushResult must write the JSON envelope to a real (non-nil) ResultWriter — the asynq queue
// path that runs in production — and Result() must keep returning the cached envelope after
// FlushResult clears the staged result (cache-before-clear invariant), here exercised against a
// non-nil writer that the Execute-path tests cannot reach.
func TestFlushResultWritesEnvelopeAndResultStaysCached(t *testing.T) {
	var buf bytes.Buffer
	ti := NewTaskInfo("t", &BasePayload{}, &buf)
	ti.WriteSuccess("ok", nil)

	if err := ti.FlushResult(); err != nil {
		t.Fatalf("FlushResult() error = %v", err)
	}

	var written TaskResult
	if err := json.Unmarshal(buf.Bytes(), &written); err != nil {
		t.Fatalf("written bytes are not a valid TaskResult envelope: %v (raw=%q)", err, buf.String())
	}

	if !written.Success || written.Message != "ok" {
		t.Errorf("written envelope = %+v, want success=true message=ok", written)
	}

	// Result() must still return the cached envelope even though FlushResult cleared _pendingResult.
	cached := ti.Result()
	if cached == nil {
		t.Fatal("Result() = nil after FlushResult, want cached envelope")
	}

	if !cached.Success || cached.Message != "ok" {
		t.Errorf("cached Result() = %+v, want success=true message=ok", cached)
	}
}
