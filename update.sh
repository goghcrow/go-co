#!/bin/bash

go get github.com/goghcrow/go-loader@main
go get github.com/goghcrow/go-matcher@main
go get github.com/goghcrow/go-imports@main
go mod tidy
git commit -am "update go.mod" && git push origin