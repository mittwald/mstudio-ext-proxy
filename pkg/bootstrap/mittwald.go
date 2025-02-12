package bootstrap

import (
	"context"
	"github.com/mittwald/api-client-go/mittwaldv2"
	generatedv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"log/slog"
)

func BuildMittwaldAPIClientFromConfig(c *Config, l *slog.Logger) generatedv2.Client {
	opts := make([]mittwaldv2.ClientOption, 0)

	if c.MittwaldBaseURL != "" {
		opts = append(opts, mittwaldv2.WithBaseURL(c.MittwaldBaseURL))
	}

	opts = append(opts, mittwaldv2.WithRequestLogging(l, c.LogHttpBodies, c.LogHttpBodies))

	client, err := mittwaldv2.New(context.Background(), opts...)
	if err != nil {
		panic(err)
	}

	return client
}
