// Package gsmanager contains the secretprovider implementation
// for GCP's secret manager.
package gsmanager

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type secretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest,
		opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

// GSManager is a secret provider that communicates with Google Cloud Platform's Secret Manager
// to decrypt secrets. Internally it uses the Google Cloud SDK.
type GSManager struct {
	smClient secretManagerClient
}

// New return's an initialized GSManager using a new secret manager client.
func New(ctx context.Context) (*GSManager, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gsmanager: failed to initialize client: %w", err)
	}

	return &GSManager{smClient: c}, nil
}

// Decrypt will access the secret on GCP Secret Manager and return the plain text string.
func (g *GSManager) Decrypt(ctx context.Context, secret string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secret,
	}

	result, err := g.smClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("gsmanager: failed to access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

// Close closes the connection to the secret manager API.
func (g *GSManager) Close() error {
	return g.smClient.Close()
}
