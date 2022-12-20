package admission

import (
	"context"

	"github.com/b4fun/tg-auth/internal/session"
)

type ReviewResult struct {
	Allowed bool
}

type Admissioner interface {
	Review(
		ctx context.Context,
		sess session.Session,
	) (ReviewResult, error)
}
