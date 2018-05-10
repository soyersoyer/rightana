#!/bin/sh
cd frontend &&
npm install &&
npm run dist &&
cd .. &&
go get github.com/rakyll/statik &&
go get ./... &&
go generate ./api &&
go build
