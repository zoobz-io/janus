package labels

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
)

func TestMain(m *testing.M) {
	// sum.NewStore requires the sum service to be initialized.
	sum.New()
	sum.Start()
	os.Exit(m.Run())
}

// memProvider is an in-memory grub.StoreProvider so the mapping logic can be
// driven without Redis.
type memProvider struct {
	data map[string][]byte
}

func newMemProvider() *memProvider { return &memProvider{data: map[string][]byte{}} }

func (p *memProvider) Get(_ context.Context, key string) ([]byte, error) {
	v, ok := p.data[key]
	if !ok {
		return nil, grub.ErrNotFound
	}
	return v, nil
}
func (p *memProvider) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	p.data[key] = value
	return nil
}
func (p *memProvider) Delete(_ context.Context, key string) error {
	if _, ok := p.data[key]; !ok {
		return grub.ErrNotFound
	}
	delete(p.data, key)
	return nil
}
func (p *memProvider) Exists(_ context.Context, key string) (bool, error) {
	_, ok := p.data[key]
	return ok, nil
}
func (p *memProvider) List(_ context.Context, prefix string, limit int) ([]string, error) {
	var keys []string
	for k := range p.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
			if limit > 0 && len(keys) >= limit {
				break
			}
		}
	}
	return keys, nil
}
func (p *memProvider) GetBatch(_ context.Context, keys []string) (map[string][]byte, error) {
	out := make(map[string][]byte, len(keys))
	for _, k := range keys {
		if v, ok := p.data[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}
func (p *memProvider) SetBatch(_ context.Context, items map[string][]byte, _ time.Duration) error {
	for k, v := range items {
		p.data[k] = v
	}
	return nil
}

// fakeApps is the read-repair / reconcile source.
type fakeApps struct {
	byID map[string]*models.Application
}

func (f *fakeApps) GetApplication(_ context.Context, id string) (*models.Application, error) {
	a, ok := f.byID[id]
	if !ok {
		return nil, fmt.Errorf("not found: %s", id)
	}
	return a, nil
}
func (f *fakeApps) ListAll(_ context.Context) ([]*models.Application, error) {
	out := make([]*models.Application, 0, len(f.byID))
	for _, a := range f.byID {
		out = append(out, a)
	}
	return out, nil
}

var storeCounter int

// newTestLabels builds an ApplicationLabels over the in-memory provider. Each
// call uses a unique catalog name (sum.NewStore panics on duplicates).
func newTestLabels(apps applicationSource) *ApplicationLabels {
	storeCounter++
	return newApplicationLabels(newMemProvider(), apps, fmt.Sprintf("test-app-labels-%d", storeCounter))
}

func TestPutAndResolve(t *testing.T) {
	ctx := context.Background()
	l := newTestLabels(&fakeApps{byID: map[string]*models.Application{}})

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
}

func TestRenameRetainsLegacyName(t *testing.T) {
	ctx := context.Background()
	l := newTestLabels(&fakeApps{byID: map[string]*models.Application{}})

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
	l := newTestLabels(&fakeApps{byID: map[string]*models.Application{}})

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
	l := newTestLabels(apps)

	// Cold cache: resolves via read-repair.
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
	l := newTestLabels(&fakeApps{byID: map[string]*models.Application{}})
	_ = l.Put(ctx, "app-6", "Dup")

	names, err := l.ResolveNames(ctx, []string{"app-6", "app-6", ""})
	if err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	if len(names) != 1 || names["app-6"] != "Dup" {
		t.Fatalf("dedupe/empty handling = %v", names)
	}
}

func TestReconcileUpsertsAll(t *testing.T) {
	ctx := context.Background()
	apps := &fakeApps{byID: map[string]*models.Application{
		"app-7": {ID: "app-7", Name: "One"},
		"app-8": {ID: "app-8", Name: "Two"},
	}}
	l := newTestLabels(apps)

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
