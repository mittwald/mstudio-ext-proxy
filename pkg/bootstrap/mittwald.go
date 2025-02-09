package bootstrap

import (
	"context"
	"github.com/mittwald/api-client-go/mittwaldv2"
	generatedv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
)

func BuildMittwaldAPIClientFromConfig(c *Config) generatedv2.Client {
	return BuildMittwaldAPIClient(context.Background(), c.MittwaldBaseURL)
}

func BuildMittwaldAPIClient(ctx context.Context, baseURL string) generatedv2.Client {
	opts := make([]mittwaldv2.ClientOption, 0)

	if baseURL != "" {
		opts = append(opts, mittwaldv2.WithBaseURL(baseURL))
	}

	client, err := mittwaldv2.New(ctx, opts...)
	if err != nil {
		panic(err)
	}

	return client
}
