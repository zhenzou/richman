#!/bin/bash

cd "$(dirname "$0")"

echo "current directory :$(pwd)}"

ln -snf $(pwd)/hooks/pre-commit $(pwd)/.git/hooks/pre-commit
chmod +x $(pwd)/.git/hooks/pre-commit


ln -snf $(pwd)/hooks/pre-push $(pwd)/.git/hooks/pre-push
chmod +x $(pwd)/.git/hooks/pre-push

chmod +x build/*


