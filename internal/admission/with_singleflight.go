package admission

import (
	"context"
	"sync"

	"github.com/b4fun/tg-auth/internal/session"
)

type singleFlightResult struct {
	Result ReviewResult
	Err    error
}

type withSingleFlight struct {
	next Admissioner

	singleFlightLock *sync.Mutex
	singleFlights    map[string][]chan singleFlightResult
}

func WithSingleFlight(
	admissioner Admissioner,
) (Admissioner, error) {
	rv := &withSingleFlight{
		next: admissioner,

		singleFlightLock: &sync.Mutex{},
		singleFlights:    make(map[string][]chan singleFlightResult),
	}

	return rv, nil
}

func (wi *withSingleFlight) Review(
	ctx context.Context,
	sess session.Session,
) (ReviewResult, error) {

	key := sess.UserID
	resultChan := make(chan singleFlightResult)

	wi.singleFlightLock.Lock()
	if chans, ok := wi.singleFlights[key]; !ok || len(chans) < 1 {
		wi.singleFlights[key] = []chan singleFlightResult{resultChan}
		go func(ctx context.Context, sess session.Session, key string) {
			reviewResult, err := wi.next.Review(ctx, sess)
			result := singleFlightResult{reviewResult, err}

			wi.singleFlightLock.Lock()
			defer wi.singleFlightLock.Unlock()
			for _, c := range wi.singleFlights[key] {
				c <- result
			}
			delete(wi.singleFlights, key)
		}(ctx, sess, key)
	} else {
		wi.singleFlights[key] = append(wi.singleFlights[key], resultChan)
	}
	wi.singleFlightLock.Unlock()

	select {
	case <-ctx.Done():
		return ReviewResult{}, ctx.Err()
	case result := <-resultChan:
		return result.Result, result.Err
	}
}
