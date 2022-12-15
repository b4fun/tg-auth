package tgauth

import "time"

type AuthnSettings struct {
	SessionTTL time.Duration `env:"AUTHN_SESSION_TTL" envDefault:"1h"`
	CookieName string        `env:"AUTHN_COOKIE_NAME"`
}

type AuthzSettings struct {
	ChannelIDs []string `env:"AUTHZ_CHANNEL_IDS" envSeparator:","`
}

type LoginSettings struct {
	RedirectURL string `env:"LOGIN_REDIRECT_URL"`
}

type BotSettings struct {
	Name  string `env:"BOT_NAME"`
	Token string `env:"BOT_TOKEN"`
}

type Settings struct {
	Login LoginSettings
	Bot   BotSettings
}
