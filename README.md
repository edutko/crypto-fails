# crypto-fails

A moderately realistic encrypted file storage/sharing app full of cryptographic vulnerabilities

## Included vulnerabilities

* CBC bit-flipping on encrypted session cookie
* CBC padding oracle on encrypted session cookie
* Weak HMAC secret key for signing API tokens
* `"alg": "none"` is accepted in API tokens
* JWT header injection via `jwk` or `jku` in API tokens
* Algorithm confusion (RS256 vs HS256) in API token signature verification
* Non-canonical encoding in session (cookie/API token) revocation list
* Nonce reuse in CTR mode for encrypted files
* Unauthenticated ciphertext (CTR mode) in encrypted files
* Length-extension on SHA-256 MAC in sharing links

## Special routes to enable attacks

To simulate attackers with the ability to read or modify ciphertext, the app provides two special
routes that are not exposed anywhere in the UI:

* `GET /vulns/leak/{key...}` exposes the entire hierarchy of encrypted files for browsing and
download
* `PUT /vulns/tweak/{key...}` accepts a file upload and will create or overwrite the corresponding
encrypted file (the filename in the upload is ignored)

Setting the environment variable `LEAK_ENCRYPTED_FILES=0` or `TWEAK_ENCRYPTED_FILES=0` disables the
respective route.

## API

Since the UI is written using <a href="https://templ.guide/">TEMPL</a>, it can be difficult to
interact with the app from a script or programming language. Almost all of the functionality
available in the UI is available in REST-ish APIs, which accept and return JSON.
