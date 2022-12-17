package session

import (
	"crypto/rand"
	"fmt"
)

const (
	unauthenticatedUserID = ""
	sessionIDChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Session struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func (sess *Session) Authenticated() bool {
	return sess.UserID != unauthenticatedUserID
}

func newSessionID() (string, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return "", fmt.Errorf("read random data: %w", err)
	}

	for i, b := range data {
		data[i] = sessionIDChars[b%byte(len(sessionIDChars))]
	}

	return string(data), nil
}

func Default() Session {
	sessionID, err := newSessionID()
	if err != nil {
		panic(fmt.Sprintf("unable to create session id: %s", err))
	}

	return Session{
		ID:     sessionID,
		UserID: unauthenticatedUserID,
	}
}
