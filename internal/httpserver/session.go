package httpserver

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/b4fun/tg-auth/internal/session"
	"github.com/b4fun/tg-auth/internal/settings"
)

type cookieSession struct {
	ExpiredAtUnixTime int64           `json:"expired_at"`
	Session           session.Session `json:"session"`
}

type cookieSigner interface {
	Encode(session *cookieSession) (string, error)
	Decode(raw string) (*cookieSession, error)
}

type aesCookieSigner struct {
	signingKey []byte
}

func newAESCookieSigner(signingKeyEncoded string) (*aesCookieSigner, error) {
	signingKey, err := base64.StdEncoding.DecodeString(signingKeyEncoded)
	if err != nil {
		return nil, fmt.Errorf("decode signing key: %w", err)
	}

	rv := &aesCookieSigner{signingKey: signingKey}
	if _, err := rv.createAEAD(); err != nil {
		// preflight check
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}

	return rv, nil
}

var _ cookieSigner = (*aesCookieSigner)(nil)

func (cs *aesCookieSigner) createAEAD() (cipher.AEAD, error) {
	block, err := aes.NewCipher(cs.signingKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return gcm, nil
}

func (cs *aesCookieSigner) Encode(session *cookieSession) (string, error) {
	encodedSession, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}

	gcm, err := cs.createAEAD()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}

	// nonce:ct
	cipherTextWithNonce := gcm.Seal(nonce, nonce, encodedSession, nil)

	return base64.StdEncoding.EncodeToString(cipherTextWithNonce), nil
}

func (cs *aesCookieSigner) Decode(raw string) (*cookieSession, error) {
	cipherTextWithNonce, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted text: %w", err)
	}

	gcm, err := cs.createAEAD()
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherTextWithNonce) < nonceSize {
		return nil, errors.New("invalid cipher text length")
	}

	nonce, cipherText := cipherTextWithNonce[:nonceSize], cipherTextWithNonce[nonceSize:]
	encodedSession, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	rv := new(cookieSession)
	if err := json.Unmarshal(encodedSession, rv); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}

	return rv, nil
}

type cookieSessionManager struct {
	cookieDomain string
	cookieName   string
	cookieSigner cookieSigner
	sessionTTL   time.Duration
}

func newCookieSessionManager(settings settings.AuthnSettings) (*cookieSessionManager, error) {
	signer, err := newAESCookieSigner(settings.CookieSigningKey)
	if err != nil {
		return nil, err
	}

	rv := &cookieSessionManager{
		cookieDomain: settings.CookieDomain,
		cookieName:   settings.CookieName,
		cookieSigner: signer,
		sessionTTL:   settings.SessionTTL,
	}
	if rv.cookieDomain == "" {
		return nil, fmt.Errorf("cookie domain is empty")
	}
	if rv.cookieName == "" {
		return nil, fmt.Errorf("cookie name is empty")
	}
	if rv.sessionTTL <= 0 {
		return nil, fmt.Errorf("negative session TTL set")
	}

	return rv, nil
}

var _ SessionManager = (*cookieSessionManager)(nil)

func (sm *cookieSessionManager) GetSession(
	ctx context.Context,
	req *http.Request,
) (session.Session, error) {
	cookie, err := req.Cookie(sm.cookieName)
	if err != nil {
		return session.Default(), fmt.Errorf("get cookie: %w", err)
	}

	cookieSession, err := sm.cookieSigner.Decode(cookie.Value)
	if err != nil {
		return session.Default(), fmt.Errorf("decode session from cookie: %w", err)
	}

	expiredAt := time.Unix(cookieSession.ExpiredAtUnixTime, 0)
	if expiredAt.Before(time.Now()) {
		// session has expired
		return session.Default(), errSessionExpired
	}

	return cookieSession.Session, nil
}

func (sm *cookieSessionManager) SetSession(
	ctx context.Context,
	rw http.ResponseWriter,
	session session.Session,
) error {
	sessionExpireAt := time.Now().Add(sm.sessionTTL)
	cookieSession := &cookieSession{
		ExpiredAtUnixTime: sessionExpireAt.Unix(),
		Session:           session,
	}

	sessionEncrypted, err := sm.cookieSigner.Encode(cookieSession)
	if err != nil {
		return fmt.Errorf("encrypt session: %w", err)
	}

	cookie := &http.Cookie{
		Name:    sm.cookieName,
		Value:   sessionEncrypted,
		Expires: sessionExpireAt,
		Secure:  true, // https only
		Domain:  sm.cookieDomain,
		Path:    "/", // allow in all paths under this domain
	}

	http.SetCookie(rw, cookie)

	return nil
}
