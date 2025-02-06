package webhooksv1

import (
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhookscommon"
)

type ExtensionInstanceSecretRotated struct {
	webhookscommon.Envelope

	ID      string  `json:"id"`
	Context Context `json:"context"`
	Meta    Meta    `json:"meta"`
	Secret  string  `json:"secret"`
}
