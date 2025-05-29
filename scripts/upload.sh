#!/usr/bin/env bash

server=http://localhost:8080

#curl -s -F username=alice -F password=password $server/login
alice_token=$(curl -s --json '{ "username": "alice", "password": "password" }' $server/api/login | jq -r '.token')
