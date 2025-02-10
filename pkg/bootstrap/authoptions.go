package bootstrap

import (
	"github.com/mittwald/mstudio-ext-proxy/pkg/authentication"
	"time"
)

func BuildAuthenticationOptions(c *Config) authentication.Options {
	if c.Secret == "" {
		panic("MITTWALD_EXT_PROXY_SECRET must be set")
	}

	return authentication.Options{
		CookieName:     "mstudio_ext_session",
		CookieTTL:      60 * time.Minute,
		JWTSecret:      []byte(c.Secret),
		StaticPassword: c.StaticPassword,
	}
}
