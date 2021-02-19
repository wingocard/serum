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
		expectedErr       error
	}{
		{
			name: "error initilazing secret provider",
			newSecretProvider: func() (secretprovider.SecretProvider, error) {
				return nil, errors.New("bad secret provider")
			},
			expectedErr: errors.New("error initializing secret provider"),
		},
		{
			name: "success",
			newSecretProvider: func() (secretprovider.SecretProvider, error) {
				return &testSecretProvider{}, nil
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ij := &Injector{}

			err := WithSecretProviderFunc(tc.newSecretProvider)(ij)

			if tc.expectedErr == nil {
				assert.NilError(t, err)
				assert.Assert(t, ij.secretProvider != nil)
				return
			}

			assert.Assert(t, ij.secretProvider == nil)
			assert.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}
