module github.com/goghcrow/go-co

go 1.19

require (
	github.com/goghcrow/go-ast-matcher v0.1.3
	github.com/goghcrow/go-imports v0.0.3-0.20240221114019-5a6ed41cc3b5
	github.com/goghcrow/go-loader v0.0.4-0.20240221113906-cab11067771f
	github.com/goghcrow/go-matcher v0.0.5-0.20240221112341-6675288f4167
	golang.org/x/tools v0.18.0
)

require golang.org/x/mod v0.15.0 // indirect

//replace github.com/goghcrow/go-matcher => ./../go-matcher
//replace github.com/goghcrow/go-loader => ./../go-loader
//replace github.com/goghcrow/go-ast-matcher => ./../go-ast-matcher
//replace github.com/goghcrow/go-imports => ./../go-imports
