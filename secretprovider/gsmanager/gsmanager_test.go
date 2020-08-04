package gsmanager

import (
	"context"
	"testing"

	"github.com/googleapis/gax-go/v2"
	. "github.com/onsi/gomega"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
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
	g := NewWithT(t)

	secretIdentifier := "my/super/secret/versions/latest"
	decrypted := []byte("superSecret")
	tc := &testClient{
		accessSecretVersionReturn: &secretmanagerpb.AccessSecretVersionResponse{
			Name: secretIdentifier,
			Payload: &secretmanagerpb.SecretPayload{
				Data: decrypted,
			},
		},
	}
	gsm := &GSManager{
		smClient: tc,
	}

	dec, err := gsm.Decrypt(context.Background(), secretIdentifier)
	g.Expect(err).To(BeNil())
	g.Expect(tc.accessSecretVersionCalled).To(BeTrue())
	g.Expect(dec).To(Equal(string(decrypted)))
}

func TestClose(t *testing.T) {
	g := NewWithT(t)

	tc := &testClient{}
	gsm := &GSManager{
		smClient: tc,
	}

	err := gsm.Close()
	g.Expect(err).To(BeNil())
	g.Expect(tc.closeCalled).To(BeTrue())
}
