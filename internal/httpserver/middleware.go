package httpserver

import (
	"net/http"

	"github.com/b4fun/tg-auth/internal/admission"
	"github.com/b4fun/tg-auth/internal/settings"
	"go.uber.org/zap"
)

func RequireAuth(
	logger *zap.Logger,
	signinSettings settings.SigninSettings,
	admissioner admission.Admissioner,
	sessionManager SessionManager,
) (MiddlewareFunc, error) {
	rootLogger := logger.Named("require-auth")
	redirectToLogin := http.RedirectHandler(signinSettings.SigninURL, http.StatusFound)

	rv := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			logger := rootLogger

			sess, err := sessionManager.GetSession(ctx, req)
			if err != nil {
				logger.Error("get session failed", zap.Error(err))
				redirectToLogin.ServeHTTP(wr, req)
				return
			}

			logger = logger.With(zap.String("session.id", sess.ID))

			reviewResult, err := admissioner.Review(ctx, sess)
			if err != nil {
				logger.Error("session admission review failed", zap.Error(err))
				redirectToLogin.ServeHTTP(wr, req)
				return
			}

			if reviewResult.Allowed {
				logger.Debug("session accepted")
				next.ServeHTTP(wr, req)
				return
			}

			logger.Debug("session rejected")
			redirectToLogin.ServeHTTP(wr, req)
		})
	}

	return rv, nil
}
