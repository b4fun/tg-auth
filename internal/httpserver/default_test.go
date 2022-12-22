package httpserver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/b4fun/tg-auth/internal/admission"
	"github.com/b4fun/tg-auth/internal/admission/testhelper"
	"github.com/b4fun/tg-auth/internal/session"
	"github.com/b4fun/tg-auth/internal/settings"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_Default_SigninFlow(t *testing.T) {
	const (
		expectedUserID         = "user-id-allowed"
		hostName               = "example.com"
		cookieName             = "tg-auth"
		signinPath             = "/signin"
		signinCallbackPath     = "/signin/callback"
		afterSigninPath        = "/"
		botToken               = "bot-token"
		defaultHandlerRespBody = "default"
	)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	var botTokenSHA256 []byte
	{
		hasher := sha256.New()
		_, err := hasher.Write([]byte(botToken))
		assert.NoError(t, err)
		botTokenSHA256 = hasher.Sum(nil)
	}

	admissioner := &testhelper.Admissioner{
		ReviewFunc: func(
			ctx context.Context,
			sess session.Session,
		) (admission.ReviewResult, error) {
			return admission.ReviewResult{
				Allowed: sess.UserID == expectedUserID,
			}, nil
		},
	}

	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(defaultHandlerRespBody))
		w.WriteHeader(http.StatusOK)
	})

	handler, err := Default(
		logger,
		settings.Settings{
			Authn: settings.AuthnSettings{
				SessionTTL:       10 * time.Minute,
				CookieDomain:     hostName,
				CookieName:       cookieName,
				CookieSigningKey: newSigningKey(t, 32),
			},
			Signin: settings.SigninSettings{
				RedirectCallbackURL: signinCallbackPath,
				SigninURL:           signinPath,
				AfterSigninURL:      afterSigninPath,
			},
			Bot: settings.BotSettings{
				Name:  "test-bot",
				Token: botToken,
			},
		},
		admissioner,
		defaultHandler,
	)
	assert.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	tc := ts.Client()
	tc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// disable redirect as we want to check the redirect response
		return http.ErrUseLastResponse
	}

	t.Log("redirect to / without auth")
	{
		resp, err := tc.Get(ts.URL + "/")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		redirectToUrl := resp.Header.Get("Location")
		assert.Equal(t, signinPath, redirectToUrl)
	}

	t.Log("request /signin without auth")
	{
		resp, err := tc.Get(ts.URL + signinPath)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "response signin html")
	}

	t.Log("request /signin/callback with expected user")
	var signinCookieValue string
	{
		qs := url.Values{}
		qs.Set(signinParameterID, expectedUserID)
		signature, err := sha256HMACSignature(botTokenSHA256, qs, func(s string) bool { return true })
		assert.NoError(t, err)
		qs.Set(signinParameterHash, signature)
		resp, err := tc.Get(fmt.Sprintf("%s%s?%s", ts.URL, signinCallbackPath, qs.Encode()))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
		redirectToUrl := resp.Header.Get("Location")
		assert.Equal(t, afterSigninPath, redirectToUrl)

		for _, respCookie := range resp.Cookies() {
			if respCookie.Name == cookieName {
				assert.Equal(t, hostName, respCookie.Domain)
				assert.Equal(t, "/", respCookie.Path)
				assert.Equal(t, true, respCookie.Secure)
				signinCookieValue = respCookie.Value
			}
		}
		assert.NotEmpty(t, signinCookieValue, "should return sign in cookie")
	}

	t.Log("request / with auth")
	{
		req := httptest.NewRequest(http.MethodGet, ts.URL+afterSigninPath, nil)
		req.RequestURI = ""
		req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: signinCookieValue,
		})
		resp, err := tc.Do(req)
		assert.NoError(t, err)

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, defaultHandlerRespBody, string(respBody))
	}

	t.Log("request /signin/callback without expected user")
	{
		qs := url.Values{}
		qs.Set(signinParameterID, "another-user-id")
		signature, err := sha256HMACSignature(botTokenSHA256, qs, func(s string) bool { return true })
		assert.NoError(t, err)
		qs.Set(signinParameterHash, signature)
		resp, err := tc.Get(fmt.Sprintf("%s%s?%s", ts.URL, signinCallbackPath, qs.Encode()))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		redirectToUrl := resp.Header.Get("Location")
		assert.Equal(t, signinPath, redirectToUrl)
	}
}
