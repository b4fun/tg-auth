package admission

import (
	"context"
	"time"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/jellydator/ttlcache/v3"
)

type withCache struct {
	next  Admissioner
	cache *ttlcache.Cache[string, ReviewResult]
}

var _ Admissioner = (*withCache)(nil)

func WithCache(
	admissioner Admissioner,
	cacheTTL time.Duration,
) (Admissioner, error) {
	rv := &withCache{
		next: admissioner,
		cache: ttlcache.New(
			ttlcache.WithTTL[string, ReviewResult](cacheTTL),
			ttlcache.WithDisableTouchOnHit[string, ReviewResult](),
		),
	}

	go rv.cache.Start()

	return rv, nil
}

func (wc *withCache) Review(
	ctx context.Context,
	sess session.Session,
) (ReviewResult, error) {
	cacheKey := sess.UserID
	cached := wc.cache.Get(cacheKey)
	if cached != nil && !cached.IsExpired() {
		return cached.Value(), nil
	}

	result, err := wc.next.Review(ctx, sess)
	if err != nil {
		return result, err
	}

	wc.cache.Set(cacheKey, result, ttlcache.DefaultTTL)
	return result, nil
}
