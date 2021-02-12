package serum

import (
	"errors"
	"fmt"

	"github.com/wingocard/serum/secretprovider"
)

// Option represents a function that can be passed into a NewInjector to
// modify its behavior.
type Option func(ij *Injector) error

// WithSecretProvider is a middleware that assigns a secretprovider.SecretProvider
// to an injector.
func WithSecretProvider(sp secretprovider.SecretProvider, err error) Option {
	return func(ij *Injector) error {
		if err != nil {
			return fmt.Errorf("error initializing secret provider: %w", err)
		}
		if sp == nil {
			return errors.New("secret provider cannot be nil")
		}

		ij.SecretProvider = sp
		return nil
	}
}
