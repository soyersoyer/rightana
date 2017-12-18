#!/bin/sh
cd frontend &&
npm install &&
npm run dist &&
cd .. &&
go get github.com/jteeuwen/go-bindata/... &&
go get github.com/elazarl/go-bindata-assetfs/... &&
go get ./... &&
go generate ./api &&
go build
