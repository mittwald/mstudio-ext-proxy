package bootstrap

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/mittwald/mstudio-ext-proxy/pkg/proxy"
)

type Config struct {
	MongoDBURI                string `envconfig:"mongodb_uri"`
	Secret                    string `required:"true"`
	StaticPassword            string `envconfig:"static_password"`
	MittwaldBaseURL           string `envconfig:"api_base_url"`
	Context                   string
	Upstreams                 proxy.ConfigurationCollection
	RedirectOnUnauthenticated string `envconfig:"redirect_on_unauthenticated"`
	LogHttpBodies             bool   `envconfig:"log_http_bodies"`
}

func ConfigFromEnv() *Config {
	c := Config{}
	envconfig.MustProcess("mittwald_ext_proxy", &c)
	return &c
}
