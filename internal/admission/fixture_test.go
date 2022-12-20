package admission

import (
	"context"

	"github.com/b4fun/tg-auth/internal/session"
)

type testAdmissionerT struct {
	ReviewFunc func(ctx context.Context, sess session.Session) (ReviewResult, error)
}

var _ Admissioner = (*testAdmissionerT)(nil)

func (ta *testAdmissionerT) Review(
	ctx context.Context,
	sess session.Session,
) (ReviewResult, error) {
	return ta.ReviewFunc(ctx, sess)
}
