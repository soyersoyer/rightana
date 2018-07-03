#!/bin/sh
cd frontend &&
npm install &&
npm run dist &&
cd .. &&
go get github.com/rakyll/statik &&
go generate ./api &&
go get ./... &&
go build
