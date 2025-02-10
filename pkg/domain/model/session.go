package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"time"
)

var _ jwt.Claims = &SessionClaims{}

type Session struct {
	ID            string `bson:"_id"`
	Expires       time.Time
	SessionSecret []byte
	UserID        string
	FirstName     string
	LastName      string
	Email         string
	AccessToken   string
	Instance      ExtensionInstance
}

func NewSession() (Session, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return Session{}, err
	}

	return Session{
		ID:            uuid.Must(uuid.NewUUID()).String(),
		SessionSecret: secret,
	}, nil
}

type SessionClaims struct {
	Session  Session
	IssuedAt time.Time
}

func (s Session) CookieString() string {
	return fmt.Sprintf("%s:%X", s.ID, s.SessionSecret)
}

func (s Session) IssueClaims() *SessionClaims {
	return &SessionClaims{
		Session:  s,
		IssuedAt: time.Now(),
	}
}

func SessionIDAndSecretFromCookieString(cookieString string) (string, []byte) {
	cookieString, err := url.QueryUnescape(cookieString)
	if err != nil {
		return "", nil
	}

	parts := strings.SplitN(cookieString, ":", 2)
	if len(parts) != 2 {
		return "", nil
	}

	secret, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", nil
	}

	return parts[0], secret
}

func (s *SessionClaims) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"exp":   s.Session.Expires.Unix(),
		"iat":   s.IssuedAt.Unix(),
		"nbf":   s.IssuedAt.Unix(),
		"iss":   "mstudio-ext-proxy",
		"sub":   s.Session.UserID,
		"fname": s.Session.FirstName,
		"lname": s.Session.LastName,
		"email": s.Session.Email,
		"inst":  s.Session.Instance,
		"tok":   s.Session.AccessToken,
	}

	return json.Marshal(out)
}

func (s SessionClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(s.Session.Expires), nil
}

func (s SessionClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(s.IssuedAt), nil
}

func (s SessionClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(s.IssuedAt), nil
}

func (s SessionClaims) GetIssuer() (string, error) {
	return "mstudio-ext-proxy", nil
}

func (s SessionClaims) GetSubject() (string, error) {
	return s.Session.UserID, nil
}

func (s SessionClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
