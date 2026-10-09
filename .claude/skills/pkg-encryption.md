---
name: Package - encryption (Cryptographic Utilities)
description: This skill should be used when the user asks to "create encryption utilities", "add hashing", "generate secrets", "add Argon2 hashing", or mentions anything about the pkg/encryption/ package.
version: 1.0.0
---

# Package: encryption — Cryptographic Utilities

The `encryption` package provides cryptographic utilities for secret generation and string hashing using Argon2 (memory-hard password hashing).

**Location:** `pkg/encryption/`

---

## Functions

### GenerateSecret

Generates a cryptographically secure random secret of the specified byte length, returned as a base64-encoded string:

```go
func GenerateSecret(length int) (string, error) {
    secret := make([]byte, length)
    _, err := rand.Read(secret)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(secret), nil
}
```

### HashString

Hashes a string using Argon2id (memory-hard, resistant to GPU/ASIC attacks) with a random salt, returning a base64-encoded hash:

```go
func HashString(value string) (string, error) {
    salt := make([]byte, 16)
    _, err := rand.Read(salt)
    if err != nil {
        return "", err
    }

    hash := argon2.IDKey([]byte(value), salt, 1, 64*1024, 4, 32)
    return base64.StdEncoding.EncodeToString(hash), nil
}
```

---

## Key Rules

- **No internal imports** — standalone package
- **Uses crypto/rand** — cryptographically secure random generation
- **Argon2id** — memory-hard hashing algorithm (not bcrypt or SHA)
- **Base64 encoding** — all outputs are base64-encoded strings
- **Copy-ready** — only depends on Go stdlib and `golang.org/x/crypto/argon2`

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
