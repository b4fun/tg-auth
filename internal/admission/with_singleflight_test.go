package admission

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/stretchr/testify/assert"
)

func Test_withSingleFlight(t *testing.T) {
	innerCalledTimes := 0

	inner := &testAdmissionerT{
		ReviewFunc: func(ctx context.Context, sess session.Session) (ReviewResult, error) {
			innerCalledTimes += 1

			time.Sleep(100 * time.Millisecond)

			return ReviewResult{}, nil
		},
	}

	withSingleFlight, err := WithSingleFlight(inner)
	assert.NoError(t, err)

	ctx := context.Background()
	sess := session.Default()
	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result, err := withSingleFlight.Review(ctx, sess)
			assert.NoError(t, err)
			assert.False(t, result.Allowed)

		}()
	}

	wg.Wait()

	assert.Equal(t, 1, innerCalledTimes)
}
