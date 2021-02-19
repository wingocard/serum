package serum

import (
	"fmt"

	"github.com/wingocard/serum/secretprovider"
)

// Option represents a function that can be passed into a NewInjector to
// modify its behavior.
type Option func(ij *Injector) error

// WithSecretProviderFunc is a middleware that accepts a function
// which returns a SecretProvider that will be assign to an injector. Any
// error returned will cause the middleware to fail and return the error.
func WithSecretProviderFunc(f func() (secretprovider.SecretProvider, error)) Option {
	return func(ij *Injector) error {
		sp, err := f()
		if err != nil {
			return fmt.Errorf("error initializing secret provider: %w", err)
		}

		ij.secretProvider = sp
		return nil
	}
}
