#!/bin/sh
cd frontend &&
npm install &&
npm run dist &&
cd .. &&
go get &&
go generate ./api &&
go build
