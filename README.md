# crypto-fails

A moderately realistic encrypted file storage/sharing app full of cryptographic vulnerabilities

## Included vulnerabilities

* CBC bit-flipping on encrypted session cookie
* CBC padding oracle on encrypted session cookie
* Weak HMAC secret key for signing API tokens
* Accepts `"alg": "none"` in API token
* JWT header injection via `jwk` or `jku` in API token
* (TODO) JWT header injection via `kid` path traversal in API token
* Algorithm confusion (RS256 vs HS256) in API token
* Non-canonical encoding in API token revocation list
* Nonce reuse in CTR mode in encrypted files
* Unauthenticated ciphertext (CTR mode) in encrypted files
* Length-extension on SHA-256 MAC in sharing links
