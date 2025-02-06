package webhookscommon

import (
	"context"
	"crypto/ed25519"
)

type KeyProvider interface {
	PublicKeyForSerial(context.Context, string) (ed25519.PublicKey, error)
}
