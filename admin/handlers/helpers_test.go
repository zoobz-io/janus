package handlers

import (
	"testing"

	"github.com/zoobz-io/rocco"
)

// requireScope asserts that an endpoint declares exactly one required scope. The
// admin surface uses single-scope endpoints (no OR/AND groups), so this pins the
// per-endpoint operator scope set by WithScopes.
func requireScope(t *testing.T, ep rocco.Endpoint, want string) {
	t.Helper()
	sg := ep.Spec().ScopeGroups
	if len(sg) != 1 || len(sg[0]) != 1 || sg[0][0] != want {
		t.Fatalf("endpoint scopes = %v, want a single [%q]", sg, want)
	}
}

func TestPathID(t *testing.T) {
	p := &rocco.Params{Path: map[string]string{"app_id": "a-1"}}
	if got := pathID(p, "app_id"); got != "a-1" {
		t.Fatalf("pathID = %q, want a-1", got)
	}
	if got := pathID(p, "missing"); got != "" {
		t.Fatalf("pathID(missing) = %q, want empty", got)
	}
}
