#!/usr/bin/env bash

server=http://localhost:8080

curl -s -F username=alice -F password=password $server/register
curl -s --json '{ "username": "bob", "password": "password" }' $server/api/register
