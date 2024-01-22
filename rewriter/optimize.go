package rewriter

import (
	"go/ast"
	"log"
	"strings"

	"github.com/goghcrow/go-ast-matcher"
	"github.com/goghcrow/go-ast-matcher/imports"
	"golang.org/x/tools/go/ast/astutil"
)

func optimize(
	inputDir, outputDir string,
	patterns []string,
	opts ...matcher.MatchOption,
) {
	opts = append(opts, matcher.WithSuppressErrors())
	m := newMatcher(inputDir, outputDir, patterns, opts...)
	seqPkg := m.All[pkgSeqPath]
	if seqPkg == nil {
		log.Printf("skip optimize: no import %s\n", pkgSeqPath)
		return
	}

	m.VisitAllFiles(func(m *matcher.Matcher, file *ast.File) {
		if !imports.Uses(m, file, seqPkg.Types) {
			log.Printf("skip file: %s\n", m.Filename)
			return
		}

		// 1. optimize file
		log.Printf("visit file: %s\n", m.Filename)
		optimizeImports(m)
		optimizeDelayCall(m)
		optimizeBindCall(m)

		// 2. write file
		log.Printf("write file: %s\n", m.Filename)
		filename := strings.ReplaceAll(m.Filename, m.Cfg.Dir, outputDir)
		m.WriteGeneratedFile(filename, pkgCoPath)
	})
}

func optimizeImports(m *matcher.Matcher) {
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
func optimizeDelayCall(m *matcher.Matcher) {
	var calleeOf func(...string) matcher.CallExprPattern
	calleeOf = func(xs ...string) matcher.CallExprPattern {
		assert(len(xs) > 0)
		callee := matcher.FuncCallee(m, pkgSeqPath, xs[0])
		if len(xs) == 1 {
			return callee
		} else {
			return matcher.Or(m, callee, calleeOf(xs[1:]...))
		}
	}

	constTrue := func(m *matcher.Matcher, n ast.Node, stack []ast.Node, binds matcher.Binds) bool { return true }

	delayCallWithReturnOnly := matcher.AndEx[matcher.CallExprPattern](m,
		matcher.FuncCallee(m, pkgSeqPath, cstDelay),
		&ast.CallExpr{
			Args: []ast.Expr{
				&ast.FuncLit{
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									matcher.Bind(m, "return",
										matcher.Or(m,
											// call seq.Delay/Combine/For/While/Loop/Range/Return
											calleeOf(cstDelay, cstCombine, cstFor, cstWhile, cstLoop, cstRange, cstReturn),

											// call seq.Bind with literal fst value
											matcher.AndEx[matcher.CallExprPattern](m,
												matcher.FuncCallee(m, pkgSeqPath, cstBind),
												&ast.CallExpr{
													Args: []ast.Expr{
														matcher.MkPattern[matcher.BasicLitPattern](m, constTrue),
														matcher.Wildcard[matcher.ExprPattern](m),
													},
												},
											),
										),
									),
								},
							},
						},
					},
				},
			},
		},
	)

	m.Match(
		delayCallWithReturnOnly,
		func(m *matcher.Matcher, c *astutil.Cursor, stack []ast.Node, binds matcher.Binds) {
			c.Replace(binds["return"])
		},
	)
}

func optimizeBindCall(m *matcher.Matcher) {
	bindCallWithReturnOnly := matcher.AndEx[matcher.CallExprPattern](m,
		matcher.FuncCallee(m, pkgSeqPath, cstBind),
		&ast.CallExpr{
			Args: []ast.Expr{
				matcher.Wildcard[matcher.ExprPattern](m),
				&ast.FuncLit{
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun:  matcher.MkVar[matcher.ExprPattern](m, "fun"),
										Args: []ast.Expr{},
									},
								},
							},
						},
					},
				},
			},
		},
	)
	m.Match(
		bindCallWithReturnOnly,
		func(m *matcher.Matcher, c *astutil.Cursor, stack []ast.Node, binds matcher.Binds) {
			bindCall := c.Node()
			bindCall.(*ast.CallExpr).Args[1] = binds["fun"].(ast.Expr)
			c.Replace(bindCall)
		},
	)
}
