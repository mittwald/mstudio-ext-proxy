package webhookscommon

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http"
)

type Verifier struct {
	KeyProvider KeyProvider
}

func (m *Verifier) VerifyWebhookRequest(ctx context.Context, request *http.Request, body []byte) error {
	serial := request.Header.Get("X-Marketplace-Signature-Serial")
	algo := request.Header.Get("X-Marketplace-Signature-Algorithm")
	signature := request.Header.Get("X-Marketplace-Signature")

	publicKey, err := m.KeyProvider.PublicKeyForSerial(ctx, serial)
	if err != nil {
		return fmt.Errorf("error retrieving public key for serial '%s': %w", serial, err)
	}

	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("error decoding signature: %w", err)
	}

	switch algo {
	case "Ed25519":
		if !ed25519.Verify(publicKey, body, decodedSignature) {
			return fmt.Errorf("invalid signature")
		}
	default:
		return fmt.Errorf("invalid signature algorithm '%s'", algo)
	}

	return nil
}
