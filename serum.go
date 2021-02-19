// Package serum allows you to inject secrets and environment variables at runtime.
package serum

import (
	"context"
	"fmt"
	"os"

	"github.com/wingocard/serum/internal/envparser"
	"github.com/wingocard/serum/secretprovider"
)

// Injector injects environment variables into the current running process. Key/value pairs can
// be read in from a .env file using the load method.
type Injector struct {
	secretProvider secretprovider.SecretProvider
	envVars        *envparser.EnvVars
}

// NewInjector creates a new injector loading from the provided loader
// and applying the provided options.
func NewInjector(loader Loader, options ...Option) (*Injector, error) {
	ij := &Injector{}
	if err := loader.Load(ij); err != nil {
		return nil, fmt.Errorf("serum: %s", err)
	}

	for _, option := range options {
		if err := option(ij); err != nil {
			return nil, fmt.Errorf("serum: %s", err)
		}
	}

	return ij, nil
}

// Inject will inject the loaded environment variables into the current running process' environment.
// Any secret values found will attempt to be decrypted using the provided SecretProvider.
// The presence of secrets with a nil SecretProvider will return an error.
func (ij *Injector) Inject(ctx context.Context) error {
	if len(ij.envVars.Secrets) > 0 && ij.secretProvider == nil {
		return fmt.Errorf("serum: error injecting env vars: secrets were loaded but the SecretProvider is nil")
	}

	// inject secrets
	for k, v := range ij.envVars.Secrets {
		decrypted, err := ij.secretProvider.Decrypt(ctx, v)
		if err != nil {
			return fmt.Errorf("serum: error decrypting secret %s: %s", v, err)
		}

		if err := os.Setenv(k, decrypted); err != nil {
			return fmt.Errorf("serum: error setting env var %s: %s", k, err)
		}
	}

	// inject plain text vars
	for k, v := range ij.envVars.Plain {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("serum: error setting env var %s: %s", k, err)
		}
	}
	return nil
}

// Close will close any open clients in the Injector.
func (ij *Injector) Close() error {
	if ij.secretProvider == nil {
		return nil
	}

	return ij.secretProvider.Close()
}
