package secretprovider

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

//GSManager is a secret provider that communicates with Google Cloud Platform's Secret Manager
//to decrypt secrets. Internally it uses the Google Cloud SDK.
type GSManager struct {
	smClient *secretmanager.Client
}

//NewGSManager return's an initialized GSManager using a new secret manager client.
func NewGSManager() (*GSManager, error) {
	c, err := secretmanager.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("gsmanager: failed to initialize client: %w", err)
	}

	return &GSManager{smClient: c}, nil
}

//Decrypt will access the secret on GCP Secret Manager and return the plain text string.
func (g *GSManager) Decrypt(secret string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secret,
	}

	result, err := g.smClient.AccessSecretVersion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("gsmanager: failed to access secret version: %w", err)
	}

	return result.Payload.String(), nil
}

//Close closes the connection to the secret manager API.
func (g *GSManager) Close() error {
	return g.smClient.Close()
}
