package testhelper

import (
	"context"

	"github.com/b4fun/tg-auth/internal/admission"
	"github.com/b4fun/tg-auth/internal/session"
)

type Admissioner struct {
	ReviewFunc func(ctx context.Context, sess session.Session) (admission.ReviewResult, error)
}

var _ admission.Admissioner = (*Admissioner)(nil)

func (ta *Admissioner) Review(
	ctx context.Context,
	sess session.Session,
) (admission.ReviewResult, error) {
	return ta.ReviewFunc(ctx, sess)
}
