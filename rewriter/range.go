package rewriter

import (
	"flag"
	"go/ast"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

var underTestRun = flag.Lookup("test.v")

func (r *yieldRewriter) gensym(prefix string) string {
	if underTestRun == nil {
		return prefix
	}
	r.symCnt++
	return prefix + strconv.Itoa(r.symCnt)
}

func (r *yieldRewriter) ignoreKeyVal(k, v ast.Expr) (bool, bool) {
	ignore := func(e ast.Expr) bool { return isNil(e) || isUnderline(e) }
	return ignore(k), ignore(v)
}

func (r *yieldRewriter) rewriteRanges(block *ast.BlockStmt) {
	do := func(c *astutil.Cursor, n *ast.RangeStmt, ctor string, arg ast.Expr) {
		factory := r.SeqSelect(ctor)
		iter := X.Call(factory, arg)
		init, forStmt := r.rewriteRangeToForIter(n, iter)
		c.InsertBefore(init)
		c.Replace(forStmt)
	}

	astutil.Apply(block, nil, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.RangeStmt:
			ty := r.rewriter.TypeOf(n.X)
			r.assert(!isNil(ty), n.X, "type missing")
			ty = ty.Underlying()

			switch ty := ty.(type) {
			case *types.Basic:
				switch {
				case ty.Info()&types.IsString != 0:
					do(c, n, cstNewStringIter, n.X)
				case ty.Info()&types.IsInteger != 0:
					// >= 1.22 only, but no release, need test
					do(c, n, cstNewIntegerIter, n.X)
				}
			case *types.Array:
				// typing workaround for abstract generic array iter
				typeInferedSlice := &ast.SliceExpr{X: n.X}
				do(c, n, cstNewSliceIter, typeInferedSlice)
			case *types.Slice:
				do(c, n, cstNewSliceIter, n.X)
			case *types.Map:
				do(c, n, cstNewMapIter, n.X)
			case *types.Chan:
				do(c, n, cstNewChanIter, n.X)
			case *types.Signature:
				panic("implement me: range func")
			}
		}
		return true
	})
}

func (r *yieldRewriter) rewriteRangeToForIter(
	n *ast.RangeStmt,
	iter ast.Expr,
) (
	init *ast.AssignStmt,
	forStmt *ast.ForStmt,
) {
	// r.rewriter.NewIdent()
	it := X.Ident(r.gensym(cstIterVar))
	current := X.Select(it, cstCurrent)
	next := X.Select(it, cstMoveNext)

	init = X.Define(it, iter)
	cond := X.Call(next)

	var kv *ast.AssignStmt
	ignoreKey, ignoreVal := r.ignoreKeyVal(n.Key, n.Value)
	switch {
	case ignoreKey && ignoreVal:
		// do nothing
	case ignoreVal:
		kv = X.Assign(n.Tok, n.Key, X.Select(X.Call(current), cstPairKey))
	case ignoreKey:
		kv = X.Assign(n.Tok, n.Value, X.Select(X.Call(current), cstPairVal))
	default:
		kv = X.Assign2(n.Tok,
			n.Key, n.Value,
			X.Select(X.Call(current), cstPairKey),
			X.Select(X.Call(current), cstPairVal),
		)
	}

	forStmt = X.ForStmt(
		nil,
		cond,
		nil,
		X.Block1(kv, n.Body.List...),
	)
	return
}
