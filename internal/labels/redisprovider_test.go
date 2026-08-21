package labels

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/grub"
)

func newTestProvider(t *testing.T) *redisProvider {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return newRedisProvider(client)
}

func TestRedisProviderRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newTestProvider(t)

	// Missing key -> grub.ErrNotFound.
	if _, err := p.Get(ctx, "nope"); !errors.Is(err, grub.ErrNotFound) {
		t.Fatalf("Get missing = %v, want ErrNotFound", err)
	}

	if err := p.Set(ctx, "k1", []byte("v1"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := p.Get(ctx, "k1")
	if err != nil || string(got) != "v1" {
		t.Fatalf("Get = %q,%v", got, err)
	}

	ok, err := p.Exists(ctx, "k1")
	if err != nil || !ok {
		t.Fatalf("Exists k1 = %v,%v", ok, err)
	}
	if ok, _ := p.Exists(ctx, "missing"); ok {
		t.Fatal("Exists missing should be false")
	}

	if err := p.Delete(ctx, "k1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := p.Delete(ctx, "k1"); !errors.Is(err, grub.ErrNotFound) {
		t.Fatalf("Delete missing = %v, want ErrNotFound", err)
	}
}

func TestRedisProviderBatchAndList(t *testing.T) {
	ctx := context.Background()
	p := newTestProvider(t)

	items := map[string][]byte{"p:a": []byte("1"), "p:b": []byte("2"), "other": []byte("3")}
	if err := p.SetBatch(ctx, items, 0); err != nil {
		t.Fatalf("SetBatch: %v", err)
	}
	// Empty batch is a no-op.
	if err := p.SetBatch(ctx, map[string][]byte{}, 0); err != nil {
		t.Fatalf("SetBatch empty: %v", err)
	}

	got, err := p.GetBatch(ctx, []string{"p:a", "p:b", "missing"})
	if err != nil {
		t.Fatalf("GetBatch: %v", err)
	}
	if string(got["p:a"]) != "1" || string(got["p:b"]) != "2" {
		t.Fatalf("GetBatch = %v", got)
	}
	if _, present := got["missing"]; present {
		t.Fatal("missing key should be omitted from GetBatch")
	}
	// Empty keys short-circuits.
	if out, err := p.GetBatch(ctx, nil); err != nil || len(out) != 0 {
		t.Fatalf("GetBatch empty = %v,%v", out, err)
	}

	keys, err := p.List(ctx, "p:", 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("List prefix p: = %v, want 2 keys", keys)
	}
	// Limit is honored.
	limited, err := p.List(ctx, "p:", 1)
	if err != nil || len(limited) != 1 {
		t.Fatalf("List limit 1 = %v,%v", limited, err)
	}
}
