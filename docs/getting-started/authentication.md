# Authentication

The Globus Go SDK supports multiple authentication methods using authorizers.

## Authorizers

All SDK clients require an authorizer that implements the `Authorizer` interface:

```go
type Authorizer interface {
    GetAuthorizationHeader() (string, error)
}
```

## Access Token Authorizer

The simplest method using a pre-obtained access token:

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"

token := "your-access-token"
authorizer := authorizers.NewAccessTokenAuthorizer(token)

// Use with any client
client := transfer.NewClient(authorizer)
```

### Getting an Access Token

You can obtain tokens via:

1. **Globus CLI**: `globus session show` after logging in
2. **Globus Developers Console**: [developers.globus.org](https://developers.globus.org)
3. **OAuth2 Flow**: Implement OAuth2 flow in your application

## Refresh Token Authorizer

Automatically refreshes expired tokens:

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"

authorizer := authorizers.NewRefreshTokenAuthorizer(
    "your-client-id",
    "your-client-secret",
    "your-refresh-token",
)

client := transfer.NewClient(authorizer)
```

## Client Credentials

For service-to-service authentication:

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"

authorizer := authorizers.NewClientCredentialsAuthorizer(
    "your-client-id",
    "your-client-secret",
)

client := transfer.NewClient(authorizer)
```

## OAuth2 Flow Example

Complete OAuth2 flow implementation:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
)

func main() {
    clientID := "your-client-id"
    clientSecret := "your-client-secret"
    redirectURI := "http://localhost:8080/callback"

    // Step 1: Generate authorization URL
    authURL := fmt.Sprintf(
        "https://auth.globus.org/v2/oauth2/authorize?"+
            "client_id=%s&"+
            "redirect_uri=%s&"+
            "scope=openid+profile+email&"+
            "response_type=code",
        clientID, redirectURI,
    )

    fmt.Printf("Visit this URL to authorize:\n%s\n\n", authURL)

    // Step 2: Handle callback
    authCode := make(chan string, 1)

    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        authCode <- code
        fmt.Fprintf(w, "Authorization successful! You can close this window.")
    })

    go http.ListenAndServe(":8080", nil)

    code := <-authCode

    // Step 3: Exchange code for tokens
    authClient := auth.NewClient(nil)
    ctx := context.Background()

    tokens, err := authClient.GetTokens(ctx, &auth.GetTokensRequest{
        GrantType:    "authorization_code",
        Code:         code,
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURI:  redirectURI,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Access Token: %s\n", tokens.AccessToken)
    fmt.Printf("Refresh Token: %s\n", tokens.RefreshToken)

    // Use the tokens
    authorizer := authorizers.NewAccessTokenAuthorizer(tokens.AccessToken)
    // ... use authorizer with clients
}
```

## Environment Variables

Store credentials securely:

```go
import "os"

token := os.Getenv("GLOBUS_ACCESS_TOKEN")
clientID := os.Getenv("GLOBUS_CLIENT_ID")
clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")
```

Set them in your shell:

```bash
export GLOBUS_ACCESS_TOKEN="your-token"
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-secret"
```

## Token Scopes

Different operations require different scopes:

| Scope | Description |
|-------|-------------|
| `openid profile email` | Basic identity information |
| `urn:globus:auth:scope:transfer.api.globus.org:all` | Transfer operations |
| `urn:globus:auth:scope:search.api.globus.org:all` | Search operations |
| `urn:globus:auth:scope:groups.api.globus.org:all` | Groups operations |

## Security Best Practices

### 1. Never Commit Credentials

```bash
# Add to .gitignore
.env
*.token
credentials.json
```

### 2. Use Environment Variables

```go
// Good
token := os.Getenv("GLOBUS_ACCESS_TOKEN")

// Bad
token := "hardcoded-token-123"
```

### 3. Rotate Credentials Regularly

Refresh tokens periodically, especially for long-running services.

### 4. Limit Token Scopes

Request only the scopes your application needs:

```go
scopes := []string{
    "urn:globus:auth:scope:transfer.api.globus.org:all",
}
```

### 5. Secure Token Storage

For production applications:

- Use secure key stores (AWS Secrets Manager, HashiCorp Vault)
- Encrypt tokens at rest
- Set appropriate file permissions: `chmod 600 tokens.json`

## Testing with Mock Authorizers

For unit tests:

```go
type MockAuthorizer struct {
    Token string
}

func (m *MockAuthorizer) GetAuthorizationHeader() (string, error) {
    return "Bearer " + m.Token, nil
}

// In tests
authorizer := &MockAuthorizer{Token: "test-token"}
client := transfer.NewClient(authorizer)
```

## Next Steps

- Learn about [Client Configuration](configuration.md)
- Review [Common Patterns](../guides/common-patterns.md)
- Check the [API Reference](../api/index.md)
