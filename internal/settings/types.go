package settings

import "time"

type AuthnSettings struct {
	SessionTTL       time.Duration `env:"AUTHN_SESSION_TTL" envDefault:"1h"`
	CookieDomain     string        `env:"AUTHN_COOKIE_DOMAIN"`
	CookieName       string        `env:"AUTHN_COOKIE_NAME"`
	CookieSigningKey string        `env:"AUTHN_COOKIE_SIGNING_KEY"`
}

type AuthzSettings struct {
	ChannelIDs []string      `env:"AUTHZ_CHANNEL_IDS" envSeparator:","`
	CacheTTL   time.Duration `env:"AUTHZ_CACHE" envDefault:"5m"`
}

type SigninSettings struct {
	RedirectCallbackURL string `env:"SIGNIN_REDIRECT_CALLBACK_URL"`
	SigninURL           string `env:"SIGNIN_URL"`
	AfterSigninURL      string `env:"SIGNIN_AFTER_SIGNIN_URL"`
}

type BotSettings struct {
	Name  string `env:"BOT_NAME"`
	Token string `env:"BOT_TOKEN"`
}

type Settings struct {
	Signin SigninSettings
	Bot    BotSettings
	Authn  AuthnSettings
	Authz  AuthzSettings
}
