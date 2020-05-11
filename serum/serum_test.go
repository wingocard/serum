package serum

import (
	"errors"
	"os"
	"testing"
	"wingocard/serum/internal/envparser"

	. "github.com/onsi/gomega"
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
	returnSecret string
	returnErr    error
}

func (ts *testSecretProvider) Decrypt(secret string) (string, error) {
	if ts.returnErr != nil {
		return "", ts.returnErr
	}

	return ts.returnSecret, nil
}

func (ts *testSecretProvider) Close() error {
	return nil
}

func TestInject(t *testing.T) {
	g := NewGomegaWithT(t)

	decryptedSecret := "my secret value"
	envVars := &envparser.EnvVars{
		Plain: map[string]string{
			"PLAIN": "123456",
		},
		Secrets: map[string]string{
			"SECRET": "superSecret",
		},
	}

	t.Cleanup(func() {
		if err := cleanupEnv(envVars); err != nil {
			t.Errorf("error cleaning up env: %s", err)
		}
	})

	ij := &Injector{
		envVars: envVars,
		SecretProvider: &testSecretProvider{
			returnSecret: decryptedSecret,
		},
	}

	err := ij.Inject()
	g.Expect(err).To(BeNil())

	for k, v := range envVars.Plain {
		g.Expect(os.Getenv(k)).To(Equal(v))
	}
	for k := range envVars.Secrets {
		g.Expect(os.Getenv(k)).To(Equal(decryptedSecret))
	}
}

func TestInjectEnvError(t *testing.T) {
	tt := []struct {
		name string
		env  *envparser.EnvVars
	}{
		{
			name: "plain set error",
			env: &envparser.EnvVars{
				Plain: map[string]string{
					"": "",
				},
			},
		},
		{
			name: "secret set error",
			env: &envparser.EnvVars{
				Secrets: map[string]string{
					"": "",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ij := &Injector{
				envVars: tc.env,
				SecretProvider: &testSecretProvider{
					returnSecret: "",
				},
			}

			err := ij.Inject()
			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(ContainSubstring("serum: error setting env var"))
		})
	}
}

func TestInjectNilSecretProviderError(t *testing.T) {
	g := NewGomegaWithT(t)

	envVars := &envparser.EnvVars{
		Secrets: map[string]string{
			"SECRET": "superSecret",
		},
	}

	ij := &Injector{
		envVars: envVars,
	}

	err := ij.Inject()
	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("serum: error injecting env vars: secrets were loaded but the SecretProvider is nil"))
}

func TestInjectDecryptError(t *testing.T) {
	g := NewGomegaWithT(t)

	envVars := &envparser.EnvVars{
		Secrets: map[string]string{
			"SECRET": "superSecret",
		},
	}

	ij := &Injector{
		envVars: envVars,
		SecretProvider: &testSecretProvider{
			returnErr: errors.New("decrypt failure"),
		},
	}

	err := ij.Inject()
	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("serum: error decrypting secret"))
}
