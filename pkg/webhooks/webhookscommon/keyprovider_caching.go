package webhookscommon

import (
	"context"
	"crypto/ed25519"
)

type KeyProviderCache struct {
	Inner KeyProvider
	cache map[string]ed25519.PublicKey
}

func (k *KeyProviderCache) PublicKeyForSerial(ctx context.Context, s string) (ed25519.PublicKey, error) {
	if k.cache == nil {
		k.cache = make(map[string]ed25519.PublicKey)
	}

	if existing, ok := k.cache[s]; ok {
		return existing, nil
	}

	key, err := k.Inner.PublicKeyForSerial(ctx, s)
	if err != nil {
		return nil, err
	}

	k.cache[s] = key
	return key, nil
}
