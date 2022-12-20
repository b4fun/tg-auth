package httpserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/b4fun/tg-auth/internal/admission"
	"github.com/b4fun/tg-auth/internal/session"
	"github.com/b4fun/tg-auth/internal/settings"
	"go.uber.org/zap"
)

const (
	signinParameterID   = "id"
	signinParameterHash = "hash"
)

type signinServer struct {
	logger           *zap.Logger
	botTokenSHA256   []byte
	admissioner      admission.Admissioner
	redirectToSignin http.Handler
	redirectToTarget http.Handler
	sessionManager   SessionManager
}

func SigninCallback(
	logger *zap.Logger,
	signinSettings settings.SigninSettings,
	botSettings settings.BotSettings,
	admissioner admission.Admissioner,
	sessionManager SessionManager,
) (*signinServer, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(botSettings.Token))
	if err != nil {
		return nil, fmt.Errorf("hash bot token: %w", err)
	}

	rv := &signinServer{
		logger:           logger.Named("signin-callback"),
		botTokenSHA256:   hasher.Sum(nil),
		admissioner:      admissioner,
		redirectToSignin: http.RedirectHandler(signinSettings.SigninURL, http.StatusFound),
		redirectToTarget: http.RedirectHandler(signinSettings.AfterSigninURL, http.StatusSeeOther),
		sessionManager:   sessionManager,
	}
	return rv, nil
}

func (ss *signinServer) validateSignin(req *http.Request) (session.Session, error) {
	sess := session.Default()

	qs := req.URL.Query()
	hash := qs.Get(signinParameterHash)
	var ps []string
	for key := range qs {
		if key == signinParameterHash {
			continue
		}
		ps = append(ps, fmt.Sprintf("%s=%s", key, qs.Get(key)))
	}
	sort.Strings(ps)
	checkStr := strings.Join(ps, "\n")

	mac := hmac.New(sha256.New, ss.botTokenSHA256)
	if _, err := mac.Write([]byte(checkStr)); err != nil {
		return sess, fmt.Errorf("calculate HMAC SHA256: %w", err)
	}
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedHash), []byte(hash)) {
		return sess, errors.New("HMAC SHA256 hash mismatch")
	}

	userID := qs.Get(signinParameterID)
	if userID == "" {
		return sess, errors.New("no user id found")
	}

	sess.UserID = userID

	return sess, nil
}

func (ss *signinServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	logger := ss.logger

	sess, err := ss.validateSignin(req)
	if err != nil {
		logger.Error("sigin validation failed", zap.Error(err))
		ss.redirectToSignin.ServeHTTP(wr, req)
		return
	}

	logger = logger.With(zap.String("session.id", sess.ID))

	ctx := req.Context()
	reviewResult, err := ss.admissioner.Review(ctx, sess)
	if err != nil {
		logger.Error("review session failed", zap.Error(err))
		ss.redirectToSignin.ServeHTTP(wr, req)
		return
	}

	if reviewResult.Allowed {
		logger.Debug("session accepted")
		if err := ss.sessionManager.SetSession(ctx, wr, sess); err != nil {
			logger.Error("set session failed", zap.Error(err))
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
		ss.redirectToTarget.ServeHTTP(wr, req)
		return
	}

	logger.Debug("session rejected")
	ss.redirectToSignin.ServeHTTP(wr, req)
}

func SigninFrontend(
	signinSettings settings.SigninSettings,
	botSettings settings.BotSettings,
) (http.Handler, error) {
	page := fmt.Sprintf(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Hello</title>
  </head>
  <body></body>
  <script async src="https://telegram.org/js/telegram-widget.js?21" data-telegram-login="%s" data-size="large" data-auth-url="%s"></script>
</html>
`, botSettings.Name, signinSettings.RedirectCallbackURL)

	rv := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, page)
	})

	return rv, nil
}
