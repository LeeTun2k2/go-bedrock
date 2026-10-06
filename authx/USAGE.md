# Using authx

## Overview

`authx` owns authentication mechanisms, while consumer services own identity storage, business policy, and authorization.

## Google (OpenID Connect)

```go
import (
    "log"

    "github.com/leetun2k2/go-bedrock/authx/provider/google"
)

googleProvider, err := google.New(google.Config{
    ClientID:     clientID,
    ClientSecret: clientSecret,
})
if err != nil {
    log.Fatalf("create google provider: %v", err)
}

loginURL, err := googleProvider.AuthorizationURL(google.AuthorizationRequest{
    RedirectURI: callbackURL,
    State:       state,
    Nonce:       nonce, // required for Google
})
if err != nil {
    log.Fatalf("build authorization url: %v", err)
}

identity, err := googleProvider.Exchange(ctx, google.ExchangeRequest{
    Code:        code,
    RedirectURI: callbackURL,
    Nonce:       nonce,
})
if err != nil {
    log.Printf("google exchange failed: %v", err)
    return
}
```

## GitHub (OAuth)

```go
import (
    "log"

    "github.com/leetun2k2/go-bedrock/authx/provider/github"
)

githubProvider, err := github.New(github.Config{
    ClientID:     clientID,
    ClientSecret: clientSecret,
})
if err != nil {
    log.Fatalf("create github provider: %v", err)
}

loginURL, err := githubProvider.AuthorizationURL(github.AuthorizationRequest{
    RedirectURI: callbackURL,
    State:       state,
})
if err != nil {
    log.Fatalf("build authorization url: %v", err)
}

identity, err := githubProvider.Exchange(ctx, github.ExchangeRequest{
    Code:        code,
    RedirectURI: callbackURL,
})
if err != nil {
    // Returned when GitHub has no verified primary email, among other proof failures.
    log.Printf("github exchange failed: %v", err)
    return
}
```

## Session Token Management (ES256)

```go
import (
    "log"
    "time"

    "github.com/leetun2k2/go-bedrock/authx/token"
)

tokens, err := token.NewManager(token.Config{
    Issuer: issuer,
    ActiveKey: token.Key{
        ID:         keyID,
        PrivateKey: privateKey,
    },
    VerificationKeys: previousKeys, // for key rotation overlap
})
if err != nil {
    log.Fatalf("create token manager: %v", err)
}

signed, err := tokens.Sign(token.Claims{
    Audience:  clientID,
    Subject:   userID,
    ExpiresAt: time.Now().Add(time.Hour).Unix(),
    Custom: map[string]any{
        "iid": identityID,
        "sid": sessionID,
    },
})
if err != nil {
    log.Fatalf("sign token: %v", err)
}

// expectedAudience is required: Verify rejects tokens issued for another audience.
claims, err := tokens.Verify(signed, clientID)
if err != nil {
    log.Printf("verify token: %v", err)
    return
}
```

## Secrets and HTTP Credentials

```go
import (
    "log"

    "github.com/leetun2k2/go-bedrock/authx/credential"
)

hash, err := credential.HashSecret(rawSecret)
if err != nil {
    log.Fatalf("hash secret: %v", err)
}

if err := credential.VerifyClientSecret(rawSecret, hash); err != nil {
    log.Printf("invalid client secret: %v", err)
    return
}

creds, err := credential.Extract(r)
if err != nil {
    log.Printf("extract credentials: %v", err)
    return
}

bearerToken, ok := credential.ExtractBearerToken(r.Header.Get("Authorization"))
if !ok {
    log.Println("missing or malformed bearer token")
    return
}
```
