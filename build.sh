#!/usr/bin/env bash

git config --global url."https://user:${TOKEN}@github.com".insteadOf "https://github.com"

make build
