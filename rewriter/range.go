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
					factory := r.SeqSelect(cstNewStringIter)
					init, forStmt := r.rewriteRangeToForIter(n, factory)
					c.InsertBefore(init)
					c.Replace(forStmt)
				case ty.Info()&types.IsInteger != 0:
					panic("implement me")
				}
			case *types.Array:
				panic("implement me")
				// factory := X.Index(
				// 	r.SeqSelect(cstNewArrayIter),
				// 	X.Any(), // ty.Elem()
				// )
				// init, forStmt := r.rewriteRangeToForIter(n, factory)
				// c.InsertBefore(init)
				// c.Replace(forStmt)
			case *types.Slice:
				// X.Index(
				// 	r.SeqSelect(cstNewSliceIter),
				// 	ty.Elem().String(), // todo type parameter
				// ),
				factory := r.SeqSelect(cstNewSliceIter)
				init, forStmt := r.rewriteRangeToForIter(n, factory)
				c.InsertBefore(init)
				c.Replace(forStmt)
			case *types.Map:
				// X.Indices(
				// 	r.SeqSelect(cstNewSliceIter),
				// 	ty.Key(),
				// 	ty.Elem(),
				// ),
				factory := r.SeqSelect(cstNewMapIter)
				init, forStmt := r.rewriteRangeToForIter(n, factory)
				c.InsertBefore(init)
				c.Replace(forStmt)
			case *types.Chan:
				panic("implement me")
			case *types.Signature:
				panic("implement me: range func")
			}
		}
		return true
	})
}

func (r *yieldRewriter) rewriteRangeToForIter(
	n *ast.RangeStmt,
	iterFactory ast.Expr,
) (
	init *ast.AssignStmt,
	forStmt *ast.ForStmt,
) {
	// r.rewriter.NewIdent()
	iter := X.Ident(r.gensym(cstIterVar))
	current := X.Select(iter, cstCurrent)
	next := X.Select(iter, cstMoveNext)

	init = X.Define(
		iter,
		X.Call(iterFactory, n.X),
	)
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
