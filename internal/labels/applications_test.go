package labels

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/events"
)

func TestMain(m *testing.M) {
	// Sync mode makes event emission deterministic; must precede default init.
	capitan.Configure(capitan.WithSyncMode())
	// sum.NewStore requires the sum service to be initialized.
	sum.New()
	sum.Start()
	os.Exit(m.Run())
}

// fakeApps is the read-repair / reconcile source.
type fakeApps struct {
	byID    map[string]*models.Application
	listErr error
}

func (f *fakeApps) GetApplication(_ context.Context, id string) (*models.Application, error) {
	a, ok := f.byID[id]
	if !ok {
		return nil, fmt.Errorf("not found: %s", id)
	}
	return a, nil
}

func (f *fakeApps) ListAll(_ context.Context) ([]*models.Application, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]*models.Application, 0, len(f.byID))
	for _, a := range f.byID {
		out = append(out, a)
	}
	return out, nil
}

var storeCounter int

// newTestLabels builds an ApplicationLabels over a real (in-process) Redis via
// miniredis, exercising the actual redis provider — no Docker required. Each call
// uses a unique catalog name (sum.NewStore panics on duplicates).
func newTestLabels(t *testing.T, apps applicationSource) *ApplicationLabels {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	storeCounter++
	return newApplicationLabels(newRedisProvider(client), apps, fmt.Sprintf("test-app-labels-%d", storeCounter))
}

func emptyApps() *fakeApps { return &fakeApps{byID: map[string]*models.Application{}} }

func TestPutAndResolve(t *testing.T) {
	ctx := context.Background()
	l := newTestLabels(t, emptyApps())

	if err := l.Put(ctx, "app-1", "Acme"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	names, err := l.ResolveNames(ctx, []string{"app-1"})
	if err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	if names["app-1"] != "Acme" {
		t.Fatalf("id->name = %q, want Acme", names["app-1"])
	}
	if id, ok, _ := l.ResolveID(ctx, "Acme"); !ok || id != "app-1" {
		t.Fatalf("name->id = %q,%v; want app-1,true", id, ok)
	}
	// Unknown name resolves to not-ok, no error.
	if _, ok, err := l.ResolveID(ctx, "Nope"); ok || err != nil {
		t.Fatalf("unknown name = ok:%v err:%v; want false,nil", ok, err)
	}
}

func TestRenameRetainsLegacyName(t *testing.T) {
	ctx := context.Background()
	l := newTestLabels(t, emptyApps())

	_ = l.Put(ctx, "app-2", "OldName")
	_ = l.Put(ctx, "app-2", "NewName")

	names, _ := l.ResolveNames(ctx, []string{"app-2"})
	if names["app-2"] != "NewName" {
		t.Fatalf("forward mapping = %q, want NewName", names["app-2"])
	}
	if id, ok, _ := l.ResolveID(ctx, "NewName"); !ok || id != "app-2" {
		t.Fatalf("NewName->id = %q,%v", id, ok)
	}
	if id, ok, _ := l.ResolveID(ctx, "OldName"); !ok || id != "app-2" {
		t.Fatalf("legacy OldName->id = %q,%v; want app-2,true", id, ok)
	}
}

func TestNameReusePointsAtNewApp(t *testing.T) {
	ctx := context.Background()
	l := newTestLabels(t, emptyApps())

	_ = l.Put(ctx, "app-3", "Shared")
	_ = l.Put(ctx, "app-3", "Renamed")
	_ = l.Put(ctx, "app-4", "Shared")

	if id, ok, _ := l.ResolveID(ctx, "Shared"); !ok || id != "app-4" {
		t.Fatalf("Shared->id = %q,%v; want app-4,true", id, ok)
	}
}

func TestReadRepairHealsColdCache(t *testing.T) {
	ctx := context.Background()
	apps := &fakeApps{byID: map[string]*models.Application{
		"app-5": {ID: "app-5", Name: "Repairable"},
	}}
	l := newTestLabels(t, apps)

	names, err := l.ResolveNames(ctx, []string{"app-5"})
	if err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	if names["app-5"] != "Repairable" {
		t.Fatalf("read-repair = %q, want Repairable", names["app-5"])
	}
	// Heal wrote both directions: reverse now resolves from cache.
	if id, ok, _ := l.ResolveID(ctx, "Repairable"); !ok || id != "app-5" {
		t.Fatalf("healed name->id = %q,%v", id, ok)
	}
	// Unknown id is omitted, not an error.
	out, err := l.ResolveNames(ctx, []string{"missing"})
	if err != nil {
		t.Fatalf("ResolveNames unknown: %v", err)
	}
	if _, present := out["missing"]; present {
		t.Fatal("unknown id should be omitted")
	}
}

func TestResolveNamesDedupesAndSkipsEmpty(t *testing.T) {
	ctx := context.Background()
	l := newTestLabels(t, emptyApps())
	_ = l.Put(ctx, "app-6", "Dup")

	names, err := l.ResolveNames(ctx, []string{"app-6", "app-6", ""})
	if err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	if len(names) != 1 || names["app-6"] != "Dup" {
		t.Fatalf("dedupe/empty handling = %v", names)
	}
	// Empty id slice short-circuits.
	if out, err := l.ResolveNames(ctx, nil); err != nil || len(out) != 0 {
		t.Fatalf("empty resolve = %v,%v", out, err)
	}
}

func TestReconcileUpsertsAll(t *testing.T) {
	ctx := context.Background()
	apps := &fakeApps{byID: map[string]*models.Application{
		"app-7": {ID: "app-7", Name: "One"},
		"app-8": {ID: "app-8", Name: "Two"},
	}}
	l := newTestLabels(t, apps)

	if err := l.Reconcile(ctx); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	names, _ := l.ResolveNames(ctx, []string{"app-7", "app-8"})
	if names["app-7"] != "One" || names["app-8"] != "Two" {
		t.Fatalf("reconcile mappings = %v", names)
	}
	// Idempotent.
	if err := l.Reconcile(ctx); err != nil {
		t.Fatalf("Reconcile 2: %v", err)
	}
}

func TestStartReconcilesAndAppliesEvents(t *testing.T) {
	ctx := context.Background()
	apps := &fakeApps{byID: map[string]*models.Application{
		"seed": {ID: "seed", Name: "Seed"},
	}}
	l := newTestLabels(t, apps)

	stop, err := l.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer stop()

	// Reconcile ran during Start.
	if id, ok, _ := l.ResolveID(ctx, "Seed"); !ok || id != "seed" {
		t.Fatal("Start did not reconcile the seed application")
	}

	// Created event is applied through the subscriber (sync mode -> deterministic).
	events.ApplicationCreated.Emit(ctx, events.ApplicationEvent{ApplicationID: "e1", Name: "Evented"})
	if id, ok, _ := l.ResolveID(ctx, "Evented"); !ok || id != "e1" {
		t.Fatal("ApplicationCreated not applied by listener")
	}

	// Updated (rename) event is applied.
	events.ApplicationUpdated.Emit(ctx, events.ApplicationEvent{ApplicationID: "e1", Name: "Renamed"})
	names, _ := l.ResolveNames(ctx, []string{"e1"})
	if names["e1"] != "Renamed" {
		t.Fatalf("ApplicationUpdated not applied: %v", names)
	}
}

func TestStartReturnsReconcileError(t *testing.T) {
	l := newTestLabels(t, &fakeApps{listErr: fmt.Errorf("db down")})

	stop, err := l.Start(context.Background())
	if err == nil {
		t.Fatal("expected reconcile error from Start")
	}
	if stop != nil {
		t.Fatal("stop should be nil when Start fails")
	}
}
