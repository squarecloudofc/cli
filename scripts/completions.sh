#!/bin/sh
set -e

rm -rf completions
mkdir completions
for sh in bash zsh fish powershell; do
	go run cmd/squarecloud/main.go completion "$sh" >"completions/completions-$sh.$sh"
done