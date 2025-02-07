package bootstrap

import (
	"github.com/mittwald/mstudio-ext-proxy/pkg/authentication"
	"os"
	"time"
)

func BuildAuthenticationOptions() authentication.Options {
	secret := os.Getenv("MITTWALD_EXT_PROXY_SECRET")
	if secret == "" {
		panic("MITTWALD_EXT_PROXY_SECRET must be set")
	}

	return authentication.Options{
		CookieName:     "mstudio_ext_session",
		CookieTTL:      60 * time.Minute,
		JWTSecret:      []byte(secret),
		StaticPassword: os.Getenv("MITTWALD_EXT_PROXY_STATIC_PASSWORD"),
	}
}
