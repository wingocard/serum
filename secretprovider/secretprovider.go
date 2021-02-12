// Package secretprovider contains the interface for
// creating a secret provider.
package secretprovider

import "context"

//SecretProvider is an interface that wraps the decrypt and close methods.
//Close should be called when the secret provier is no longer needed.
//It may be a no-op in cases where there's no underlying connection to be closed.
type SecretProvider interface {
	Decrypt(ctx context.Context, secret string) (string, error)
	Close() error
}
