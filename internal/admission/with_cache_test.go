package admission

import (
	"context"
	"testing"
	"time"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/stretchr/testify/assert"
)

func Test_withCache(t *testing.T) {
	innerCalledTimes := 0

	inner := &testAdmissionerT{
		ReviewFunc: func(ctx context.Context, sess session.Session) (ReviewResult, error) {
			innerCalledTimes += 1

			return ReviewResult{}, nil
		},
	}

	withCache, err := WithCache(inner, 10*time.Minute)
	assert.NoError(t, err)

	ctx := context.Background()
	sess := session.Default()

	for i := 0; i < 3; i++ {
		result, err := withCache.Review(ctx, sess)
		assert.NoError(t, err)
		assert.False(t, result.Allowed)
	}

	assert.Equal(t, 1, innerCalledTimes)
}
