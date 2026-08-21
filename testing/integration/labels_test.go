//go:build testing

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/database/models"
)

// flushLabels clears the Redis label cache so each test starts clean (cleanAll
// only truncates SQL tables).
func flushLabels(t *testing.T) {
	t.Helper()
	if err := testRedis.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}
}

func TestApplicationLabelMapping(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	flushLabels(t)

	t.Run("CreateResolvesBothDirections", func(t *testing.T) {
		if err := testAppLabels.Put(ctx, "app-1", "Acme"); err != nil {
			t.Fatalf("Put: %v", err)
		}
		names, err := testAppLabels.ResolveNames(ctx, []string{"app-1"})
		if err != nil {
			t.Fatalf("ResolveNames: %v", err)
		}
		if names["app-1"] != "Acme" {
			t.Fatalf("id->name = %q, want Acme", names["app-1"])
		}
		id, ok, err := testAppLabels.ResolveID(ctx, "Acme")
		if err != nil || !ok || id != "app-1" {
			t.Fatalf("name->id = %q,%v,%v; want app-1,true,nil", id, ok, err)
		}
	})

	t.Run("RenameUpdatesForwardAndRetainsLegacy", func(t *testing.T) {
		flushLabels(t)
		_ = testAppLabels.Put(ctx, "app-2", "OldName")
		if err := testAppLabels.Put(ctx, "app-2", "NewName"); err != nil {
			t.Fatalf("Put rename: %v", err)
		}
		// Forward mapping is current.
		names, _ := testAppLabels.ResolveNames(ctx, []string{"app-2"})
		if names["app-2"] != "NewName" {
			t.Fatalf("id->name = %q, want NewName", names["app-2"])
		}
		// New name resolves.
		if id, ok, _ := testAppLabels.ResolveID(ctx, "NewName"); !ok || id != "app-2" {
			t.Fatalf("NewName->id = %q,%v; want app-2,true", id, ok)
		}
		// Legacy name still resolves (append-only reverse direction).
		if id, ok, _ := testAppLabels.ResolveID(ctx, "OldName"); !ok || id != "app-2" {
			t.Fatalf("OldName->id = %q,%v; want app-2,true (legacy must resolve)", id, ok)
		}
	})

	t.Run("NameReusePointsAtNewApplication", func(t *testing.T) {
		flushLabels(t)
		_ = testAppLabels.Put(ctx, "app-3", "Shared")
		_ = testAppLabels.Put(ctx, "app-3", "Renamed") // app-3 vacates "Shared"
		if err := testAppLabels.Put(ctx, "app-4", "Shared"); err != nil {
			t.Fatalf("Put reuse: %v", err)
		}
		if id, ok, _ := testAppLabels.ResolveID(ctx, "Shared"); !ok || id != "app-4" {
			t.Fatalf("Shared->id = %q,%v; want app-4,true (reuse repoints)", id, ok)
		}
	})
}

func TestApplicationLabelReconcile(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	flushLabels(t)

	a, _ := testStores.Applications.CreateApplication(ctx, "Reconcile One", "recon-1")
	b, _ := testStores.Applications.CreateApplication(ctx, "Reconcile Two", "recon-2")

	// Cache is cold; reconcile upserts every application's mapping.
	if err := testAppLabels.Reconcile(ctx); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	names, _ := testAppLabels.ResolveNames(ctx, []string{a.ID, b.ID})
	if names[a.ID] != "Reconcile One" || names[b.ID] != "Reconcile Two" {
		t.Fatalf("reconcile mappings = %v", names)
	}

	// Idempotent: a second run yields the same state.
	if err := testAppLabels.Reconcile(ctx); err != nil {
		t.Fatalf("Reconcile 2: %v", err)
	}
	if id, ok, _ := testAppLabels.ResolveID(ctx, "Reconcile One"); !ok || id != a.ID {
		t.Fatalf("post-reconcile name->id = %q,%v", id, ok)
	}
}

func TestApplicationLabelReadRepair(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	flushLabels(t)

	// Created in the DB but never pushed to the cache (no subscriber running).
	app, _ := testStores.Applications.CreateApplication(ctx, "Repairable", "repairable")

	// Cold cache: ResolveNames must read-repair from the applications store.
	names, err := testAppLabels.ResolveNames(ctx, []string{app.ID})
	if err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	if names[app.ID] != "Repairable" {
		t.Fatalf("read-repair id->name = %q, want Repairable", names[app.ID])
	}

	// The repair healed both directions: name->id now resolves from cache.
	if id, ok, _ := testAppLabels.ResolveID(ctx, "Repairable"); !ok || id != app.ID {
		t.Fatalf("healed name->id = %q,%v; want %s,true", id, ok, app.ID)
	}

	// Unknown id resolves to nothing (omitted), no error.
	empty, err := testAppLabels.ResolveNames(ctx, []string{"does-not-exist"})
	if err != nil {
		t.Fatalf("ResolveNames unknown: %v", err)
	}
	if _, present := empty["does-not-exist"]; present {
		t.Fatal("unknown id should be omitted")
	}
}

func TestScopeResponseReflectsApplicationRename(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	flushLabels(t)

	app, _ := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")
	scope := &models.Scope{ApplicationID: app.ID, Name: "read", Description: "read access"}

	// Initial label (via read-repair on a cold cache).
	resp, err := transformers.ScopeToResponse(ctx, scope, testAppLabels)
	if err != nil {
		t.Fatalf("ScopeToResponse: %v", err)
	}
	if resp.Application != "Nexus" {
		t.Fatalf("scope.application = %q, want Nexus", resp.Application)
	}

	// Rename the application and push the new mapping (as the event subscriber would).
	if _, err := testStores.Applications.Update(ctx, app.ID, "NexusCorp", models.ApplicationStatusActive); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := testAppLabels.Put(ctx, app.ID, "NexusCorp"); err != nil {
		t.Fatalf("Put rename: %v", err)
	}

	// The scope response now carries the renamed label.
	resp, err = transformers.ScopeToResponse(ctx, scope, testAppLabels)
	if err != nil {
		t.Fatalf("ScopeToResponse after rename: %v", err)
	}
	if resp.Application != "NexusCorp" {
		t.Fatalf("scope.application after rename = %q, want NexusCorp", resp.Application)
	}
}
