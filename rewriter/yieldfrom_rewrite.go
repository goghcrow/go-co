package rewriter

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

type yieldFromRewriter struct {
	rewriter *rewriter
}

func mkYieldFromRewriter(r *rewriter) *yieldFromRewriter {
	return &yieldFromRewriter{rewriter: r}
}

func (r *yieldFromRewriter) rewrite(c *astutil.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ExprStmt:
		if call, ok := r.rewriter.isYieldFromCall(n); ok {
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
	r.rewriter.updateUses(yield, r.rewriter.yieldFunc)

	// wrap the Yield type parameter the same as YieldFrom
	if idx, ok := call.Fun.(*ast.IndexExpr); ok {
		yield = X.Index(yield, idx.Index)
	}

	r.assert(len(call.Args) == 1, call, "invalid args num")
	iter := call.Args[0]
	rgIt := r.rangeIter(yield, iter)

	_, isYieldCall := r.rewriter.isYieldCall(rgIt.Body.List[0])
	r.assert(isYieldCall, iter, "illegal state")

	return rgIt
}

func (r *yieldFromRewriter) rangeIter(
	yieldFun ast.Expr,
	iter ast.Expr,
) *ast.RangeStmt {
	key := X.Ident(cstYieldFromRangeVar)
	callYield := X.Stmt(
		X.Call(yieldFun, key),
	)
	return &ast.RangeStmt{
		Key:  key,
		Tok:  token.DEFINE,
		X:    iter,
		Body: X.Block(callYield),
	}
}

func (r *yieldFromRewriter) assert(ok bool, pos ast.Node, format string, a ...any) {
	r.rewriter.assert(ok, pos, format, a...)
}
