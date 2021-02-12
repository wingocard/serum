# Serum

> NOTE: Serum is still pre v1.0, the API is still evolving and breaking changes can occur

Serum is a library that facilitates injecting environment variables and secrets into your application at run time.
It can load the key/value pairs from a `.env` file and it can use a `SecretProvider` to decrypt the secrets that
are included in the `.env` file.

This helps us solve two major different problems:

1. It allows us to use environment variables for our service configuration. Env vars are OS and language agnostic and can very easily be changed between application runs. This makes local development much easier to configure on a per machine basis and this benefit also extends to our cloud and CI/CD providers.

2. It allows us to keep secrets out of our source code. Using a `SecretProvider` allows those secrets to remain safely encrypted and only accessible from sources that are controlled via ACLs. Because of the way the `SecretProvider` works, it also allows local development to continue without worrying about having access to production/staging secrets. A developer's secret can be defined locally and never checked in to source control.

```sh
#Example .env file
KEY=value
NAME=Oberyn Martell
SECRET=!{secret-identifier}
```

### Secrets
Secrets are denoted in a `.env` file by surrounding the identifier with `!{}`.
Serum will pass this identifer to the specified `SecretProvider` for decryption. If the decryption is successful,
the value will be injected into the running process' environment using the specified key.

## Secret Stores

A list of secret stores currently supported:

- [GCP Secret Manager](https://cloud.google.com/secret-manager)
    - SecretProvider: `GSManager`


## Example usage

```go
package main

import (
    "context"
    "os"

    "github.com/wingocard/serum"
    "github.com/wingocard/serum/secretprovider/gsmanager"
)

func main() {
    // create a new Injector
    ij, err := serum.NewInjector(
        serum.FromFile("path/to/file.env"),
        serum.WithSecretProvider(gsmanager.New(context.Background())),
    )
    if err != nil {
        //...
    }

    // close SecretProvider connection when done
    defer ij.SecretProvider.Close()

    // inject the serum...
    if err := ij.Inject(context.Background()); err != nil {
        //...
    }

    // access env vars
    ev := os.Getenv("myKey")
}
```

## Running Tests

Run all tests using the Makefile:
`make tests`

## Linting

Code can be linted using golangci-lint:
`make lint`

> it can be installed using brew: `brew install golangci-lint`


