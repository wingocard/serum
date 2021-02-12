package serum

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/wingocard/serum/internal/envparser"
	"github.com/wingocard/serum/secretprovider"
	"gotest.tools/v3/assert"
)

func cleanupEnv(env *envparser.EnvVars) error {
	for k := range env.Plain {
		if err := os.Unsetenv(k); err != nil {
			return err
		}
	}
	for k := range env.Secrets {
		if err := os.Unsetenv(k); err != nil {
			return err
		}
	}

	return nil
}

type testSecretProvider struct {
	returnSecret map[string]string
	returnErr    error
}

func (ts *testSecretProvider) Decrypt(ctx context.Context, secret string) (string, error) {
	if ts.returnErr != nil {
		return "", ts.returnErr
	}

	return ts.returnSecret[secret], nil
}

func (ts *testSecretProvider) Close() error {
	return nil
}

func TestNewInjector(t *testing.T) {
	fakeLoader := func(ij *Injector) error {
		ij.envVars = &envparser.EnvVars{
			Plain: map[string]string{
				"loaded": "plain",
			},
			Secrets: map[string]string{
				"loaded": "secret",
			},
		}
		return nil
	}
	errLoader := func(ij *Injector) error {
		return errors.New("error loading")
	}

	optionZero := func(ij *Injector) error {
		ij.envVars.Plain["zero"] = "zero"
		return nil
	}
	optionOne := func(ij *Injector) error {
		ij.envVars.Plain["one"] = "one"
		return nil
	}
	errOption := func(ij *Injector) error {
		return errors.New("option error")
	}

	tt := []struct {
		name        string
		loader      Loader
		options     []Option
		expectedEnv *envparser.EnvVars
		expectedErr error
	}{
		{
			name:        "error loading",
			loader:      LoaderFunc(errLoader),
			options:     nil,
			expectedEnv: nil,
			expectedErr: errors.New("error loading"),
		},
		{
			name:    "one option",
			loader:  LoaderFunc(fakeLoader),
			options: []Option{optionZero},
			expectedEnv: &envparser.EnvVars{
				Plain: map[string]string{
					"loaded": "plain",
					"zero":   "zero",
				},
				Secrets: map[string]string{
					"loaded": "secret",
				},
			},
			expectedErr: nil,
		},
		{
			name:    "multiple options",
			loader:  LoaderFunc(fakeLoader),
			options: []Option{optionZero, optionOne},
			expectedEnv: &envparser.EnvVars{
				Plain: map[string]string{
					"loaded": "plain",
					"zero":   "zero",
					"one":    "one",
				},
				Secrets: map[string]string{
					"loaded": "secret",
				},
			},
			expectedErr: nil,
		},
		{
			name:        "error in option",
			loader:      LoaderFunc(fakeLoader),
			options:     []Option{optionZero, errOption},
			expectedErr: errors.New("option err"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ij, err := NewInjector(tc.loader, tc.options...)

			if tc.expectedErr == nil {
				assert.NilError(t, err)
				assert.DeepEqual(t, ij.envVars, tc.expectedEnv)
				return
			}

			assert.Assert(t, ij == nil)
			assert.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}

func TestInject(t *testing.T) {
	tt := []struct {
		name             string
		env              *envparser.EnvVars
		decryptedSecrets map[string]string
	}{
		{
			name: "secrets",
			env: &envparser.EnvVars{
				Secrets: map[string]string{
					"secret": "superSecret",
				},
			},
			decryptedSecrets: map[string]string{
				"superSecret": "super secret",
			},
		},
		{
			name: "plain",
			env: &envparser.EnvVars{
				Plain: map[string]string{
					"gwyn": "lord of cinder",
				},
			},
		},
		{
			name: "secrets and plain",
			env: &envparser.EnvVars{
				Plain: map[string]string{
					"gwyn": "lord of cinder",
				},
				Secrets: map[string]string{
					"sif": "greatWolf",
				},
			},
			decryptedSecrets: map[string]string{
				"greatWolf": "great wolf",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ij := &Injector{
				envVars: tc.env,
				SecretProvider: &testSecretProvider{
					returnSecret: tc.decryptedSecrets,
				},
			}

			err := ij.Inject(context.Background())
			assert.NilError(t, err)

			for k, v := range tc.env.Plain {
				assert.Equal(t, os.Getenv(k), v)
			}
			for k, v := range tc.env.Secrets {
				assert.Equal(t, os.Getenv(k), tc.decryptedSecrets[v])
			}

			if err := cleanupEnv(tc.env); err != nil {
				t.Errorf("error cleaning up env: %s", err)
			}
		})
	}
}

func TestInjectError(t *testing.T) {
	tt := []struct {
		name           string
		env            *envparser.EnvVars
		secretprovider secretprovider.SecretProvider
		expectedErr    error
	}{
		{
			name: "secrets with nil secret provider",
			env: &envparser.EnvVars{
				Secrets: map[string]string{
					"samus": "aran",
				},
			},
			secretprovider: nil,
			expectedErr:    errors.New("secrets were loaded but the SecretProvider is nil"),
		},
		{
			name: "plain set error",
			env: &envparser.EnvVars{
				Plain: map[string]string{
					"": "",
				},
			},
			secretprovider: &testSecretProvider{},
			expectedErr:    errors.New("serum: error setting env var"),
		},
		{
			name: "secret set error",
			env: &envparser.EnvVars{
				Secrets: map[string]string{
					"": "",
				},
			},
			secretprovider: &testSecretProvider{},
			expectedErr:    errors.New("serum: error setting env var"),
		},
		{
			name: "decrypt error",
			env: &envparser.EnvVars{
				Secrets: map[string]string{
					"solaire": "of astora",
				},
			},
			secretprovider: &testSecretProvider{
				returnErr: errors.New("decrypt failure"),
			},
			expectedErr: errors.New("error decrypting secret"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ij := &Injector{
				envVars:        tc.env,
				SecretProvider: tc.secretprovider,
			}

			err := ij.Inject(context.Background())
			assert.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}
