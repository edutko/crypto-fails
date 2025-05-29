#!/usr/bin/env bash

server=http://localhost:8080

alice_token=$(curl -s --json '{ "username": "alice", "password": "password" }' $server/api/login | jq -r '.token')
