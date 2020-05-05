package secretprovider

//SecretProvider is an interface that wraps the decrypt and close methods.
//Close should be called when the secret provier is no longer needed.
//It may be a no-op in cases where there's no underlying connection to be closed.
type SecretProvider interface {
	Decrypt(secret string) (string, error)
	Close() error
}
