package webhooks

import (
	"encoding/json"
	"fmt"
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhooksv1"

	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhookscommon"
)

func UnmarshalWebhookRequest(body []byte) (any, *webhookscommon.Envelope, error) {
	envelope := webhookscommon.Envelope{}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling webhook body into envelope: %w", err)
	}

	switch envelope.APIVersion {
	case "v1":
		wh, err := UnmarshalWebhookRequestV1(&envelope, body)
		return wh, &envelope, err
	default:
		return nil, &envelope, fmt.Errorf("unknown API version in webhook body: %s", envelope.APIVersion)
	}
}

func UnmarshalWebhookRequestV1(envelope *webhookscommon.Envelope, body []byte) (any, error) {
	var target any

	switch envelope.Kind {
	case "ExtensionAddedToContext":
		target = &webhooksv1.ExtensionAddedToContext{}
	case "ExtensionInstanceUpdated", "InstanceUpdated":
		target = &webhooksv1.ExtensionInstanceUpdated{}
	case "ExtensionInstanceSecretRotated", "SecretRotated":
		target = &webhooksv1.ExtensionInstanceSecretRotated{}
	case "ExtensionInstanceRemovedFromContext", "InstanceRemovedFromContext":
		target = &webhooksv1.ExtensionInstanceRemovedFromContext{}
	default:
		return nil, fmt.Errorf("unknown webhook kind %s", envelope.Kind)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return nil, fmt.Errorf("error unmarshaling webhook body into %T: %w", target, err)
	}

	return target, nil
}
