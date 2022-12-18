package admission

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/b4fun/tg-auth/internal/settings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type telegramAdmissioner struct {
	logger            *zap.Logger
	bot               *tgbotapi.BotAPI
	channelIDsToCheck []string
}

var _ Adminssioner = (*telegramAdmissioner)(nil)

func NewTelegramChannelAdmission(
	logger *zap.Logger,
	botSettings settings.BotSettings,
	authzSettings settings.AuthzSettings,
) (Adminssioner, error) {
	bot, err := tgbotapi.NewBotAPI(botSettings.Token)
	if err != nil {
		return nil, err
	}

	rv := &telegramAdmissioner{
		logger:            logger.Named("telegram-channel-admission"),
		bot:               bot,
		channelIDsToCheck: authzSettings.ChannelIDs,
	}

	return rv, nil
}

var validMembershiptatus = map[string]struct{}{
	"creator":       {},
	"administrator": {},
	"member":        {},
}

func (ta *telegramAdmissioner) checkChatMembership(
	ctx context.Context,
	channelID string,
	userID string,
) error {
	userIDInt64, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}
	channelIDInt64, err := strconv.ParseInt(channelID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse channel id: %w", err)
	}

	param := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: channelIDInt64,
			UserID: userIDInt64,
		},
	}
	cm, err := ta.bot.GetChatMember(param)
	if err != nil {
		return fmt.Errorf("get chat member (%q, %q): %w", userID, channelID, err)
	}
	memberStatus := strings.ToLower(cm.Status)
	if _, exists := validMembershiptatus[memberStatus]; !exists {
		return fmt.Errorf("chat member (%q, %q) status %q", userID, channelID, memberStatus)
	}

	return nil
}

func (ta *telegramAdmissioner) Review(
	ctx context.Context,
	sess session.Session,
) (ReviewResult, error) {
	rv := ReviewResult{Allowed: false}

	if !sess.Authenticated() {
		ta.logger.Debug("session is not authenticated, rejecting")
		rv.Allowed = false
		return rv, nil
	}

	if len(ta.channelIDsToCheck) < 1 {
		ta.logger.Warn("no channel ids to check, rejecting")
		return rv, nil
	}

	for _, channelID := range ta.channelIDsToCheck {
		if err := ta.checkChatMembership(ctx, channelID, sess.UserID); err != nil {
			ta.logger.Warn("channel membership check failed", zap.Error(err))
			rv.Allowed = false
			return rv, nil
		}
	}

	ta.logger.Debug("channel membership checks passsed")
	rv.Allowed = true

	return rv, nil
}
