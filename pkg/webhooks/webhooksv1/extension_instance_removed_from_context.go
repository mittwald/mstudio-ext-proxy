package webhooksv1

import (
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhookscommon"
)

type ExtensionInstanceRemovedFromContext struct {
	webhookscommon.Envelope

	ID              string   `json:"id"`
	Context         Context  `json:"context"`
	ConsentedScopes []string `json:"consentedScopes"`
	State           State    `json:"state"`
	Meta            Meta     `json:"meta"`
}
