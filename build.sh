#!/bin/sh
cd web &&
npm ci &&
npm run dist &&
cd .. &&
go get &&
go generate ./internal/api &&
go build
