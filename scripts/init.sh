#!/usr/bin/env bash

server="http://localhost:8080"

token=$(curl -s --json '{"username": "admin", "password": "admin"}' $server/api/login | jq -r .token)

# Upload expired license files
curl -s -H "Authorization: Bearer $token" -F 'file=@assets/license-2023.json' $server/api/files
curl -s -H "Authorization: Bearer $token" -F 'file=@assets/license-2024.json' $server/api/files

# Create some users
curl -s -H "Authorization: Bearer $token" --json '{"username": "alice", "password": "password"}' $server/api/users
curl -s -H "Authorization: Bearer $token" --json '{"username": "bob", "password": "password"}' $server/api/users
curl -s -H "Authorization: Bearer $token" --json '{"username": "eve", "password": "password"}' $server/api/users
curl -s -H "Authorization: Bearer $token" --json '{"username": "batman", "password": "password", "realName": "Bruce Wayne"}' $server/api/users
curl -s -H "Authorization: Bearer $token" --json '{"username": "catwoman", "password": "password", "realName": "Selina Kyle"}' $server/api/users
echo ""
echo "Users:"
curl -s -H "Authorization: Bearer $token" $server/api/users | jq -r '.users[]'

# Upload and share some files (Alice)
token=$(curl -s --json '{"username": "alice", "password": "password"}' $server/api/login | jq -r .token)
curl -s -H "Authorization: Bearer $token" -F 'file=@assets/2233BE76FC6CA28FA002A9F581F9F00A80C53A90.pub' $server/api/files
curl -s -H "Authorization: Bearer $token" -F 'file=@-;filename=corp-card.txt' $server/api/files <<EOF
Card number: 6011000990139424
Expires: Jan 2030
CVV: 123
EOF

# Upload Alice's public key
curl -s -H "Authorization: Bearer $token" -F 'file=@assets/2233BE76FC6CA28FA002A9F581F9F00A80C53A90.pub' $server/api/keys

curl -s -H "Authorization: Bearer $token" -F 'file=@-;filename=secrets.dat' $server/api/files <<< 'Be sure to drink your Ovaltine'
echo ""
echo "Alice's files:"
curl -s -H "Authorization: Bearer $token" $server/api/files | jq -r '.files[]|.key'
echo ""
echo "Alice's public keys:"
curl -s -H "Authorization: Bearer $token" $server/api/keys | jq -r '.keys[]'
echo ""
echo "Alice's shared file link:"
curl -s -H "Authorization: Bearer $token" --json '{ "key": "corp-card.txt" }' $server/api/shares | jq -r .url

# Upload and share some files (Bob)
token=$(curl -s --json '{"username": "bob", "password": "password"}' $server/api/login | jq -r .token)
curl -s -H "Authorization: Bearer $token" -F 'file=@assets/bob.pub' $server/api/files
curl -s -H "Authorization: Bearer $token" -F 'file=@-;filename=bobs-passwords.txt' $server/api/files <<'EOF'
"bank account","bob","password1"
"brokerage","bob@example.com","12345678"
"crypto seed phrase",,"monkey dragon love shadow master soccer"
EOF

# Upload Bob's public key
curl -s -H "Authorization: Bearer $token" -F 'file=@assets/126F3F7E5498F6B9D065B372EB8D78E708F4C7BA.pub' $server/api/keys

echo ""
echo "Bob's files:"
curl -s -H "Authorization: Bearer $token" $server/api/files | jq -r '.files[]|.key'
echo ""
echo "Bob's public keys:"
curl -s -H "Authorization: Bearer $token" $server/api/keys | jq -r '.keys[]'

echo ""
