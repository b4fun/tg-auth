package httpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/b4fun/tg-auth/internal/settings"
	"github.com/stretchr/testify/assert"
)

func newSigningKey(t *testing.T, keySize int) string {
	t.Helper()

	key := make([]byte, keySize)
	_, err := rand.Read(key)
	assert.NoError(t, err)

	return base64.StdEncoding.EncodeToString(key)
}

func Test_aesCookieSigner(t *testing.T) {
	t.Run("invalid signing key - base64 decode", func(t *testing.T) {
		_, err := newAESCookieSigner("foobar")
		assert.Error(t, err)
	})

	t.Run("invalid signing key - invalid aes key", func(t *testing.T) {
		b := base64.StdEncoding.EncodeToString([]byte("foobar"))
		_, err := newAESCookieSigner(b)
		assert.Error(t, err)
	})

	for _, keySize := range []int{16, 24, 32} {
		t.Run(fmt.Sprintf("encode - decode - key size: %d", keySize), func(t *testing.T) {
			keyEncoded := newSigningKey(t, keySize)

			signer, err := newAESCookieSigner(keyEncoded)
			assert.NoError(t, err)

			t.Run("session", func(t *testing.T) {
				sess := &cookieSession{
					ExpiredAtUnixTime: time.Now().Unix(),
					Session:           session.Default(),
				}

				encodedSession, err := signer.Encode(sess)
				assert.NoError(t, err)

				sessDecoded, err := signer.Decode(encodedSession)
				assert.NoError(t, err)

				assert.Equal(t, sess, sessDecoded)
			})

			t.Run("decode invalid input", func(t *testing.T) {
				_, err := signer.Decode("foobar")
				assert.Error(t, err)

				_, err = signer.Decode("Zm9vYmFyCg==")
				assert.Error(t, err)

				_, err = signer.Decode("Zm9vYmFyZm9vYmFyZm9vYmFyCg==")
				assert.Error(t, err)
			})
		})
	}
}

func Test_cookieSessionManager(t *testing.T) {
	settings := settings.AuthnSettings{
		SessionTTL:       time.Hour,
		CookieDomain:     "example.com",
		CookieName:       "auth",
		CookieSigningKey: newSigningKey(t, 32),
	}
	manager, err := newCookieSessionManager(settings)
	assert.NoError(t, err)

	t.Run("blank", func(t *testing.T) {
		_, err := manager.GetSession(context.Background(), &http.Request{})
		assert.Error(t, err)
	})

	t.Run("SetSession", func(t *testing.T) {
		sess := session.Default()
		sess.UserID = "test-user-id"

		rw := httptest.NewRecorder()
		err := manager.SetSession(
			context.Background(),
			rw,
			sess,
		)
		assert.NoError(t, err)

		v := rw.Header().Get("Set-Cookie")
		assert.NotEmpty(t, v)
		assert.Contains(t, v, "auth=")
	})
}
