package rewriter

import (
	"go/ast"
	"log"
	"strings"

	. "github.com/goghcrow/go-ast-matcher"
	"github.com/goghcrow/go-ast-matcher/imports"
	"golang.org/x/tools/go/ast/astutil"
)

func optimize(
	inputDir, outputDir string,
	patterns []string,
	opts ...MatchOption,
) {
	opts = append(opts, WithSuppressErrors())
	m := newMatcher(inputDir, outputDir, patterns, opts...)
	seqPkg := m.All[pkgSeqPath]
	if seqPkg == nil {
		log.Printf("skip optimize: no import %s\n", pkgSeqPath)
		return
	}

	m.VisitAllFiles(func(m *Matcher, file *ast.File) {
		if !imports.Uses(m, file, seqPkg.Types) {
			log.Printf("skip file: %s\n", m.Filename)
			return
		}

		// 1. optimize file
		log.Printf("visit file: %s\n", m.Filename)
		optimizeImports(m)
		optimizeDelayCall(m)
		// optimizeBindCall(m)
		etaReduction(m)

		// 2. write file
		log.Printf("write file: %s\n", m.Filename)
		filename := strings.ReplaceAll(m.Filename, m.Cfg.Dir, outputDir)
		m.WriteGeneratedFile(filename, pkgCoPath)
	})
}

func optimizeImports(m *Matcher) {
	imports.Clean(m, m.File)
}

// NOTICE:
// Delay[int](func() Seq { return Bind($variable, ...) })  !=> Bind($variable, ...)
// because $variable may be updated after Bind($variable, {{ in here }})
// Transform variable to closure upValue (value -> reference) by Delay,
// so we can visit the latest value of variable everytime

// e.g.
//
//			fib := func() Iter[int] {
//				a, b := 1, 1
//				for {
//					Yield(b)
//					a, b = b, a+b
//				}
//			}
//
//	 if we optimize delay, b in the fst pos of binding call is always 1
//
//			fib := func() Iterator {
//				return Start(Delay(func() Seq {
//					a, b := 1, 1
//					return Loop(     ===Delay(func() Seq {===
//						return Bind(b, func() Seq {
//							a, b = b, a+b
//							return Normal()
//						})
//					})===     )
//				}))
//			}
func optimizeDelayCall(m *Matcher) {
	var calleeOf func(...string) CallExprPattern
	calleeOf = func(xs ...string) CallExprPattern {
		assert(len(xs) > 0)
		callee := FuncCallee(m, pkgSeqPath, xs[0])
		if len(xs) == 1 {
			return callee
		} else {
			return Or(m, callee, calleeOf(xs[1:]...))
		}
	}

	constTrue := func(m *Matcher, n ast.Node, stack []ast.Node, binds Binds) bool { return true }

	// Currently only the first parameter of Bind has side effects
	noEffectBindCall := AndEx[CallExprPattern](m,
		FuncCallee(m, pkgSeqPath, cstBind),
		// seq.Bind[T](literal, ...)
		&ast.CallExpr{
			Args: []ast.Expr{
				MkPattern[BasicLitPattern](m, constTrue), // literal
				Wildcard[ExprPattern](m),                 // whatever
			},
		},
	)
	noEffectCombinatorCall := calleeOf(
		cstDelay,
		cstCombine,
		cstFor,
		cstWhile,
		cstLoop,
		cstReturn,
	)
	delayCallWithNoEffectDirectReturn := AndEx[CallExprPattern](m,
		FuncCallee(m, pkgSeqPath, cstDelay),
		&ast.CallExpr{
			Args: []ast.Expr{
				&ast.FuncLit{
					Body: X.Block(
						X.Return(
							Bind(m, "return",
								Or(m,
									noEffectCombinatorCall, // call seq.Delay/Combine/For/While/Loop/Range/Return
									// Or
									noEffectBindCall, // call seq.Bind with literal fst value
								),
							),
						),
					),
				},
			},
		},
	)

	m.Match(
		delayCallWithNoEffectDirectReturn,
		func(m *Matcher, c *astutil.Cursor, stack []ast.Node, binds Binds) {
			c.Replace(binds["return"])
		},
	)
}

// eat reduction overrides this particularity optimization
// no longer required
func optimizeBindCall(m *Matcher) {
	// Bind[T](*, func() Seq[T] { return [Normal|Break|Continue|...]() })
	// =>
	// Bind[T](*, [Normal|Break|Continue|...]())
	bindCallWithDirectReturn := AndEx[CallExprPattern](m,
		FuncCallee(m, pkgSeqPath, cstBind),
		&ast.CallExpr{
			Args: []ast.Expr{
				Wildcard[ExprPattern](m),
				&ast.FuncLit{
					Body: X.Block(
						X.Return(
							&ast.CallExpr{
								Fun:  MkVar[ExprPattern](m, "fun"),
								Args: []ast.Expr{},
							},
						),
					),
				},
			},
		},
	)
	m.Match(
		bindCallWithDirectReturn,
		func(m *Matcher, c *astutil.Cursor, stack []ast.Node, binds Binds) {
			bindCall := c.Node()
			bindCall.(*ast.CallExpr).Args[1] = binds["fun"].(ast.Expr)
			c.Replace(bindCall)
		},
	)
}

// fun(...args) { return return f(...args) }  ==>  f
func etaReduction(m *Matcher) {
	pattern := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: MkVar[FieldListPattern](m, "params"),
		},
		Body: X.Block(
			X.Return(
				&ast.CallExpr{
					Fun:  MkVar[ExprPattern](m, "fun"),
					Args: MkVar[ExprsPattern](m, "args"),
				},
			),
		),
	}

	// assume type-checked
	matched := func(m *Matcher, paramsFields []*ast.Field, argsExprs []ast.Expr) bool {
		if paramsFields == nil && argsExprs == nil {
			return true
		}
		var args []*ast.Ident
		for _, argExpr := range argsExprs {
			arg, _ := argExpr.(*ast.Ident)
			if arg == nil {
				return false // must be ident
			}
			args = append(args, arg)
		}

		var params []*ast.Ident
		for _, paramGroup := range paramsFields {
			for _, param := range paramGroup.Names {
				params = append(params, param)
			}
		}

		if len(args) != len(params) {
			return false
		}

		for i, arg := range args {
			param := params[i]
			if arg.Name != param.Name {
				return false
			}
			if m.ObjectOf(arg) != m.ObjectOf(param) {
				return false
			}
		}

		return true
	}

	m.Match(
		pattern,
		func(m *Matcher, c *astutil.Cursor, stack []ast.Node, binds Binds) {
			params := binds["params"].(*ast.FieldList).List
			args := binds["args"].(ExprsNode)
			if matched(m, params, args) {
				c.Replace(binds["fun"])
			}
		},
	)
}
