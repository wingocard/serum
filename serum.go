package serum

import (
	"fmt"
	"os"

	"github.com/wingocard/serum/internal/envparser"
	"github.com/wingocard/serum/secretprovider"
)

// Injector injects environment variables into the current running process. Key/value pairs can
// be read in from a .env file using the load method.
type Injector struct {
	SecretProvider secretprovider.SecretProvider
	envVars        *envparser.EnvVars
}

// Inject will inject the loaded environment variables into the current running process' environment.
// Any secret values found will attempt to be decrypted using the provided secret provider.
// The presence of secrets with a nil SecretProvider will return an error.
func (in *Injector) Inject() error {
	if len(in.envVars.Secrets) > 0 && in.SecretProvider == nil {
		return fmt.Errorf("serum: error injecting env vars: secrets were loaded but the SecretProvider is nil")
	}

	// inject secrets
	for k, v := range in.envVars.Secrets {
		decrypted, err := in.SecretProvider.Decrypt(v)
		if err != nil {
			return fmt.Errorf("serum: error decrypting secret %s: %s", v, err)
		}

		if err := os.Setenv(k, decrypted); err != nil {
			return fmt.Errorf("serum: error setting env var %s: %s", k, err)
		}
	}

	// inject plain text vars
	for k, v := range in.envVars.Plain {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("serum: error setting env var %s: %s", k, err)
		}
	}
	return nil
}

// Load will parse a .env file for key/value pairs and prepair them to be injected using the
// Inject method.
func (in *Injector) Load(path string) error {
	envVars, err := envparser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("serum: error loading env vars: %s", err)
	}

	in.envVars = envVars
	return nil
}
