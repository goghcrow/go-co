package main

import (
	"os"

	"github.com/goghcrow/go-co/rewriter"
)

func main() {
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		panic("Must run in go:generate mode")
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rewriter.GoGen(cwd)
}
