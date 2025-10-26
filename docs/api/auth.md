# Auth Service

Authentication and identity management operations.

## Package

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
```

## Create Client

```go
client := auth.NewClient(authorizer)
```

## Methods

### GetUserInfo

Get information about the authenticated user:

```go
userInfo, err := client.GetUserInfo(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Username: %s\n", userInfo.PreferredUsername)
fmt.Printf("Email: %s\n", userInfo.Email)
```

### GetTokens

Exchange authorization code for tokens:

```go
tokens, err := client.GetTokens(ctx, &auth.GetTokensRequest{
    GrantType:    "authorization_code",
    Code:         authCode,
    ClientID:     clientID,
    ClientSecret: clientSecret,
    RedirectURI:  redirectURI,
})
```

### RevokeToken

Revoke an access or refresh token:

```go
err := client.RevokeToken(ctx, &auth.RevokeTokenRequest{
    Token:        token,
    ClientID:     clientID,
    ClientSecret: clientSecret,
})
```

## See Also

- [Authentication Guide](../getting-started/authentication.md)
- [pkg.go.dev API Docs](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth)
