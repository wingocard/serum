package serum

import (
	"errors"
	"testing"

	"github.com/wingocard/serum/secretprovider"
	"gotest.tools/v3/assert"
)

func TestWithSecretProvider(t *testing.T) {
	tt := []struct {
		name              string
		newSecretProvider func() (secretprovider.SecretProvider, error)
		expectErr         error
	}{
		{
			name: "success",
			newSecretProvider: func() (secretprovider.SecretProvider, error) {
				return &testSecretProvider{}, nil
			},
			expectErr: nil,
		},
		{
			name: "error",
			newSecretProvider: func() (secretprovider.SecretProvider, error) {
				return nil, errors.New("bad secret provider")
			},
			expectErr: errors.New("error initializing secret provider"),
		},
		{
			name: "nil secret provider",
			newSecretProvider: func() (secretprovider.SecretProvider, error) {
				return nil, nil
			},
			expectErr: errors.New("secret provider cannot be nil"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ij := &Injector{}

			err := WithSecretProvider(tc.newSecretProvider())(ij)

			if tc.expectErr == nil {
				assert.NilError(t, err)
				assert.Assert(t, ij.SecretProvider != nil)
				return
			}

			assert.Assert(t, ij.SecretProvider == nil)
			assert.ErrorContains(t, err, tc.expectErr.Error())
		})
	}
}
