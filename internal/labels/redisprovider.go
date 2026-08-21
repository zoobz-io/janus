// Package labels maintains id<->human-label mappings in Redis, kept current by
// domain events and reconciled from the source tables on boot. API surfaces
// carry the label in place of the raw foreign key; labels are never resolved
// via SQL joins.
package labels

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/grub"
)

// redisProvider adapts a go-redis client to grub.StoreProvider so label stores
// can use the sanctioned sum.NewStore KV path over the shared Redis client.
// grub ships no redis provider at the pinned version, so we supply this thin one.
type redisProvider struct {
	client *goredis.Client
}

func newRedisProvider(client *goredis.Client) *redisProvider {
	return &redisProvider{client: client}
}

func (p *redisProvider) Get(ctx context.Context, key string) ([]byte, error) {
	v, err := p.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, grub.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (p *redisProvider) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return p.client.Set(ctx, key, value, ttl).Err()
}

func (p *redisProvider) Delete(ctx context.Context, key string) error {
	n, err := p.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return grub.ErrNotFound
	}
	return nil
}

func (p *redisProvider) Exists(ctx context.Context, key string) (bool, error) {
	n, err := p.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (p *redisProvider) List(ctx context.Context, prefix string, limit int) ([]string, error) {
	var keys []string
	iter := p.client.Scan(ctx, 0, prefix+"*", int64(limit)).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if limit > 0 && len(keys) >= limit {
			break
		}
	}
	return keys, iter.Err()
}

func (p *redisProvider) GetBatch(ctx context.Context, keys []string) (map[string][]byte, error) {
	out := make(map[string][]byte, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	vals, err := p.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	for i, v := range vals {
		s, ok := v.(string)
		if !ok {
			continue // nil (missing) or non-string: omit
		}
		out[keys[i]] = []byte(s)
	}
	return out, nil
}

func (p *redisProvider) SetBatch(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	pipe := p.client.Pipeline()
	for k, v := range items {
		pipe.Set(ctx, k, v, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}
