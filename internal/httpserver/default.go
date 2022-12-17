package httpserver

import (
	"net/http"

	"github.com/b4fun/tg-auth/internal/admission"
	"github.com/b4fun/tg-auth/internal/settings"
	"go.uber.org/zap"
)

func Default(
	logger *zap.Logger,
	settings settings.Settings,
	admissioner admission.Adminssioner,
	defaultHandler http.Handler,
) (http.Handler, error) {
	sessionManager, err := newCookieSessionManager(settings.Authn)
	if err != nil {
		return nil, err
	}

	signinCallback, err := SigninCallback(
		logger,
		settings.Signin, settings.Bot,
		admissioner,
		sessionManager,
	)
	if err != nil {
		return nil, err
	}

	signinFrontend, err := SigninFrontend(settings.Signin, settings.Bot)
	if err != nil {
		return nil, err
	}

	requireAuth, err := RequireAuth(
		logger,
		settings.Signin,
		admissioner,
		sessionManager,
	)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/signin", signinFrontend)
	mux.Handle("/signin/callback", signinCallback)
	mux.Handle("/", requireAuth(defaultHandler))

	return mux, nil
}
