#!/bin/sh
cd web &&
npm ci &&
npm run dist &&
cd .. &&
go get github.com/rakyll/statik &&
go generate ./internal/api &&
go get &&
go build
