module github.com/goghcrow/go-co

go 1.19

require (
	github.com/goghcrow/go-ast-matcher v0.0.13-0.20240122164752-8b23ddb7d0dc
	golang.org/x/tools v0.17.0
)

require golang.org/x/mod v0.14.0 // indirect

//replace github.com/goghcrow/go-ast-matcher => ./../go-ast-matcher
