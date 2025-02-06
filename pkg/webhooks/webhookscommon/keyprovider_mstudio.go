package webhookscommon

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	mittwaldv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/marketplaceclientv2"
)

type KeyProviderMStudio struct {
	Client mittwaldv2.Client
}

func (k *KeyProviderMStudio) PublicKeyForSerial(ctx context.Context, serial string) (ed25519.PublicKey, error) {
	req := marketplaceclientv2.GetPublicKeyRequest{Serial: serial}
	resp, _, err := k.Client.Marketplace().GetPublicKey(ctx, req)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(resp.Key)
	if err != nil {
		return nil, err
	}

	return ed25519.PublicKey(decoded), nil
}
