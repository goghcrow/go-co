package rewriter

import (
	"go/ast"
	"log"

	"github.com/goghcrow/go-ast-matcher"
	"github.com/goghcrow/go-imports"
	"github.com/goghcrow/go-loader"
	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

type optimizer struct {
	m astmatcher.ASTMatcher
}

func mkOptimizer(m astmatcher.ASTMatcher) *optimizer {
	return &optimizer{m}
}

func (o *optimizer) optimizeAllFiles(printer FilePrinter) {
	seqPkg := o.m.Loader.LookupPackage(pkgSeqPath)
	if seqPkg == nil {
		log.Printf("skip optimize: no import %s\n", pkgSeqPath)
		return
	}

	o.m.Loader.VisitAllFiles(func(f *loader.File) {
		if !imports.Uses(f, seqPkg.Types) {
			log.Printf("skip file: %s\n", f.Filename)
			return
		}

		// 1. optimize file
		log.Printf("visit file: %s\n", f.Filename)
		o.optimizeImports(f)
		o.optimizeDelayCall()
		// o.optimizeBindCall()
		o.etaReduction()

		// 2. write file
		log.Printf("write file: %s\n", f.Filename)
		printer(f.Filename, f)
	})
}

func (o *optimizer) optimizeImports(f *loader.File) {
	imports.Clean(o.m.Loader, f)
}

// NOTICE:
// Combine($seq, Return()) can be optimized to $seq
// e.g.,
//  for {
//		// Combine(For(...), Return())
//		for i := 0; i < 3; i++ {
//			Yield(i)
//		}
//		return nil
//	}
// endless loop without Return()

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
func (o *optimizer) optimizeDelayCall() {
	m := o.m.Matcher

	var calleeOf func(...string) CallExprPattern
	calleeOf = func(xs ...string) CallExprPattern {
		assert(len(xs) > 0)
		qualified := pkgSeqPath + "." + xs[0]
		funObj := o.m.Loader.MustLookup(qualified)
		callee := FuncCallee(m, funObj, xs[0])
		if len(xs) == 1 {
			return callee
		} else {
			return Or(m, callee, calleeOf(xs[1:]...))
		}
	}

	constTrue := func(n ast.Node, ctx *matcher.MatchCtx) bool { return true }

	var (
		bindFnObj  = o.m.Loader.MustLookup(pkgSeqPath + "." + cstBind)
		dealyFnObj = o.m.Loader.MustLookup(pkgSeqPath + "." + cstDelay)
	)

	// Currently only the first parameter of Bind has side effects
	noEffectBindCall := AndEx[CallExprPattern](m,
		FuncCallee(m, bindFnObj, cstBind),
		// seq.Bind[T](literal, ...)
		&ast.CallExpr{
			Args: []ast.Expr{
				matcher.MkPattern[BasicLitPattern](m, constTrue), // literal
				Wildcard[ExprPattern](m),                         // whatever
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
		FuncCallee(m, dealyFnObj, cstDelay),
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

	o.m.Match(
		delayCallWithNoEffectDirectReturn,
		func(c *astmatcher.Cursor, ctx astmatcher.Ctx) {
			c.Replace(ctx.Binds["return"])
		},
	)
}

// eat reduction overrides this particularity optimization
// no longer required
func (o *optimizer) optimizeBindCall() {
	m := o.m.Matcher

	bindFnObj := o.m.Loader.MustLookup(pkgSeqPath + "." + cstBind)
	// Bind[T](*, func() Seq[T] { return [Normal|Break|Continue|...]() })
	// =>
	// Bind[T](*, [Normal|Break|Continue|...]())
	bindCallWithDirectReturn := AndEx[CallExprPattern](m,
		FuncCallee(m, bindFnObj, cstBind),
		&ast.CallExpr{
			Args: []ast.Expr{
				Wildcard[ExprPattern](m),
				&ast.FuncLit{
					Body: X.Block(
						X.Return(
							&ast.CallExpr{
								Fun:  matcher.MkVar[ExprPattern](m, "fun"),
								Args: []ast.Expr{},
							},
						),
					),
				},
			},
		},
	)
	o.m.Match(
		bindCallWithDirectReturn,
		func(c *astmatcher.Cursor, ctx astmatcher.Ctx) {
			bindCall := c.Node()
			bindCall.(*ast.CallExpr).Args[1] = ctx.Binds["fun"].(ast.Expr)
			c.Replace(bindCall)
		},
	)
}

// fun(...args) { return return f(...args) }  ==>  f
func (o *optimizer) etaReduction() {
	m := o.m.Matcher

	pattern := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: matcher.MkVar[FieldListPattern](m, "params"),
		},
		Body: X.Block(
			X.Return(
				&ast.CallExpr{
					Fun:  matcher.MkVar[ExprPattern](m, "fun"),
					Args: matcher.MkVar[ExprsPattern](m, "args"),
				},
			),
		),
	}

	// assume type-checked
	matched := func(ctx astmatcher.Ctx, paramsFields []*ast.Field, argsExprs []ast.Expr) bool {
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
			if ctx.ObjectOf(arg) != ctx.ObjectOf(param) {
				return false
			}
		}

		return true
	}

	o.m.Match(
		pattern,
		func(c *astmatcher.Cursor, ctx astmatcher.Ctx) {
			params := ctx.Binds["params"].(*ast.FieldList).List
			args := ctx.Binds["args"].(ExprsNode)
			if matched(ctx, params, args) {
				c.Replace(ctx.Binds["fun"])
			}
		},
	)
}
