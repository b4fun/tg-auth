package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/b4fun/tg-auth/internal/session"
)

var errSessionExpired = errors.New("session expired")

type SessionManager interface {
	GetSession(
		ctx context.Context,
		req *http.Request,
	) (session.Session, error)

	SetSession(
		ctx context.Context,
		rw http.ResponseWriter,
		session session.Session,
	) error
}

type MiddlewareFunc func(next http.Handler) http.Handler
