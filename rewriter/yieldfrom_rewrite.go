package rewriter

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/goghcrow/go-loader"
	"golang.org/x/tools/go/ast/astutil"
)

type yieldFromRewriter struct {
	rewriter *rewriter
	pkg      loader.Pkg
}

func mkYieldFromRewriter(r *rewriter, pkg loader.Pkg) func(*astutil.Cursor, loader.Pkg) bool {
	return (&yieldFromRewriter{rewriter: r, pkg: pkg}).rewrite
}

func (r *yieldFromRewriter) rewrite(c *astutil.Cursor, pkg loader.Pkg) bool {
	switch n := c.Node().(type) {
	case *ast.ExprStmt:
		if call, ok := r.rewriter.isYieldFromCall(r.pkg, n); ok {
			c.Replace(r.rewriteYieldFrom(call))
			return true
		}
	}
	return true
}

// YieldFrom($iter) =>
//
//	for v := range $iter {
//		Yield(v)
//	}
func (r *yieldFromRewriter) rewriteYieldFrom(call *ast.CallExpr) *ast.RangeStmt {
	yield := X.PkgSelect(r.rewriter.coImportedName, cstAPIYield)

	// We will match `Yield[T](x)` by types.Object in isYieldCall method,
	// so we need to update info.Uses[GeneratedYield] to `yieldFunc`
	// `typeutil.Callee` calling in isYieldCall use info.Uses only, not info.Selections
	// so we only update info.Uses here
	r.pkg.UpdateUses(yield, r.rewriter.yieldFunc)

	// wrap the Yield type parameter the same as YieldFrom
	if idx, ok := call.Fun.(*ast.IndexExpr); ok {
		yield = X.Index(yield, idx.Index)
	}

	iter := call.Args[0]
	iterTyArg := r.checkYieldCall(call)
	return r.rangeIter(call, yield, iter, iterTyArg)
}

func (r *yieldFromRewriter) checkYieldCall(call *ast.CallExpr) types.Type {
	r.assert(len(call.Args) == 1, call, "invalid args num")
	iter := call.Args[0]
	tyOfIt := r.pkg.TypeOf(iter) // co.Iter[V]

	msg := "invalid YieldFrom arg type"
	r.assert(instanceof[*types.Named](tyOfIt), call, msg)

	iterNamed := tyOfIt.(*types.Named)
	r.assert(iterNamed.Obj() == r.rewriter.iterType, call, msg)

	assert(iterNamed.TypeArgs().Len() == 1)
	return iterNamed.TypeArgs().At(0)
}

func (r *yieldFromRewriter) rangeIter(
	pos *ast.CallExpr,
	yieldFun ast.Expr,
	iter ast.Expr,
	iterT types.Type,
) *ast.RangeStmt {
	// key := X.Ident(cstYieldFromRangeVar)
	// make ident with a type for checkYieldCall
	key := r.pkg.NewIdent(cstYieldFromRangeVar, iterT)

	call := X.Call(yieldFun, key)
	call.Lparen = pos.Lparen
	call.Rparen = pos.Rparen

	callYield := X.Stmt(call)

	_, isYieldCall := r.rewriter.isYieldCall(r.pkg, callYield)
	assert(isYieldCall)

	return &ast.RangeStmt{
		Key:  key,
		Tok:  token.DEFINE,
		X:    iter,
		Body: X.Block(callYield),
	}
}

func (r *yieldFromRewriter) assert(ok bool, pos ast.Node, format string, a ...any) {
	r.rewriter.assert(r.pkg, ok, pos, format, a...)
}
