package bootstrap

import (
	mittwaldv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhookscommon"
)

func BuildWebhookVerifier(client mittwaldv2.Client) *webhookscommon.Verifier {
	keyProvider := webhookscommon.KeyProviderCache{Inner: &webhookscommon.KeyProviderMStudio{Client: client}}
	webhookVerifier := webhookscommon.Verifier{KeyProvider: &keyProvider}

	return &webhookVerifier
}
