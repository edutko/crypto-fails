#!/usr/bin/env bash

NAME=$1
NAME_LOWER_NOSPACE=$(echo -n "$NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '.')
EMAIL=$NAME_LOWER_NOSPACE@example.com

(
cd assets || exit 1

GNUPGHOME=$(mktemp -d)
export GNUPGHOME

gpg --batch --generate-key <<EOF
     Key-Type: RSA
     Key-Length: 2048
     Name-Real: $NAME
     Name-Email: $EMAIL
     Expire-Date: 0
     %no-protection
     %commit
EOF

keyid=$(gpg --list-keys --with-colons 2>/dev/null | grep -E '^fpr:' | cut -d ':' -f 10)
gpg --armor --export "$keyid" > "${NAME_LOWER_NOSPACE}-${keyid}.pub"

rm -rf "$GNUPGHOME" 2>/dev/null || true
)
