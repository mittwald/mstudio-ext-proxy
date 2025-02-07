package authentication

import "time"

type Options struct {
	CookieName     string
	CookieTTL      time.Duration
	JWTSecret      []byte
	StaticPassword string
}
