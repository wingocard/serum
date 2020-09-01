package gsmanager

import (
	"context"
	"testing"

	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"gotest.tools/v3/assert"
)

type testClient struct {
	accessSecretVersionCalled bool
	accessSecretReturnError   error
	accessSecretVersionReturn *secretmanagerpb.AccessSecretVersionResponse
	closeCalled               bool
}

func (tc *testClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest,
	opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	tc.accessSecretVersionCalled = true
	if tc.accessSecretReturnError != nil {
		return nil, tc.accessSecretReturnError
	}

	return tc.accessSecretVersionReturn, nil
}

func (tc *testClient) Close() error {
	tc.closeCalled = true
	return nil
}

func TestDecrypt(t *testing.T) {
	secretIdentifier := "my/super/secret/versions/latest"
	decrypted := "superSecret"
	tc := &testClient{
		accessSecretVersionReturn: &secretmanagerpb.AccessSecretVersionResponse{
			Name: secretIdentifier,
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte(decrypted),
			},
		},
	}
	gsm := &GSManager{
		smClient: tc,
	}

	dec, err := gsm.Decrypt(context.Background(), secretIdentifier)
	assert.NilError(t, err)
	assert.Equal(t, tc.accessSecretVersionCalled, true)
	assert.Equal(t, dec, decrypted)
}

func TestClose(t *testing.T) {
	tc := &testClient{}
	gsm := &GSManager{
		smClient: tc,
	}

	err := gsm.Close()
	assert.NilError(t, err)
	assert.Equal(t, tc.closeCalled, true)
}
