#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "==> Building compiler..."
go build -o playground/mon_lang .

echo "==> Building frontend..."
cd playground/web
bun install --frozen-lockfile 2>/dev/null || bun install
bun run build
cd ..

echo "==> Copying static files..."
rm -rf static
cp -r web/out static

echo "==> Building server..."
go build -o playground-server .

echo "==> Starting playground at http://localhost:8080"
exec ./playground-server -compiler ./mon_lang -stdlib ../stdlib -port 8080
