package serum

import (
	"fmt"

	"github.com/wingocard/serum/internal/envparser"
)

// A Loader loads env variables from a source into an Injector and prepares
// them to be injected using the the Inject method.
type Loader interface {
	Load(ij *Injector) error
}

// LoaderFunc is an adapter type that allows an ordinary function
// to be used as a loader.
type LoaderFunc func(ij *Injector) error

// Load implements the loader interface.
func (f LoaderFunc) Load(ij *Injector) error {
	return f(ij)
}

// FromFile returns a loader that will parse a .env file for key/value pairs and
// assign them to an Injector.
func FromFile(path string) Loader {
	return LoaderFunc(func(ij *Injector) error {
		envVars, err := envparser.ParseFile(path)
		if err != nil {
			return fmt.Errorf("error loading env vars from file: %w", err)
		}

		ij.envVars = envVars
		return nil
	})
}

// FromEnv returns a loader that will parse the current process' environment for
// the specified keys and assigns them to an Injector.
func FromEnv(keys []string) Loader {
	return LoaderFunc(func(ij *Injector) error {
		envVars, err := envparser.ParseEnv(keys)
		if err != nil {
			return fmt.Errorf("error loading env vars from env: %s", err)
		}

		ij.envVars = envVars
		return nil
	})
}
