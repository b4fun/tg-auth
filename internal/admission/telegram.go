package admission

import (
	"context"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/b4fun/tg-auth/internal/settings"
	"go.uber.org/zap"
)

type telegramAdmissioner struct {
}

var _ Adminssioner = (*telegramAdmissioner)(nil)

func NewTelegramChannelAdmission(
	logger *zap.Logger,
	botSettings settings.BotSettings,
	authzSettings settings.AuthzSettings,
) (Adminssioner, error) {
	rv := &telegramAdmissioner{}

	return rv, nil
}

func (ta *telegramAdmissioner) Review(
	ctx context.Context,
	sess session.Session,
) (ReviewResult, error) {
	return ReviewResult{Allowed: true}, nil
}
