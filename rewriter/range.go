package rewriter

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func (r *yieldRewriter) gensym(prefix string) string {
	if runningWithGoTest {
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
			do := func(ctor string, arg ast.Expr) {
				factory := r.SeqSelect(ctor)
				iter := X.Call(factory, arg)
				init, forStmt := r.rewriteRangeToForIter(n, iter)
				c.InsertBefore(init)
				c.Replace(forStmt)
			}

			ty := r.pkg.TypeOf(n.X)
			r.assert(!isNil(ty), n.X, "type missing")
			ty = ty.Underlying()

			switch ty := ty.(type) {
			case *types.Basic:
				switch {
				case ty.Info()&types.IsString != 0:
					do(cstNewStringIter, n.X)
				case ty.Info()&types.IsInteger != 0:
					// >= 1.22 only, but no release, need test
					do(cstNewIntegerIter, n.X)
				}
			case *types.Array:
				// typing workaround for abstract generic array iter
				// type can't be infered from array, so we wrap it with slice
				typeInfered := &ast.SliceExpr{X: n.X}
				do(cstNewSliceIter, typeInfered)
			case *types.Slice:
				do(cstNewSliceIter, n.X)
			case *types.Map:
				do(cstNewMapIter, n.X)
			case *types.Chan:
				do(cstNewChanIter, n.X)
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
		forStmt = X.ForStmt(nil, cond, nil, n.Body)
		return

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

	if n.Tok == token.DEFINE {
		// wrap BlockStmt, prevent from name conflicting
		// for k, v := range it { k, v :=...  }
		// 	=>
		// it := NewXXXIter(it)
		// for it.MoveNext() {
		// 		k,v := it.Current().Key, it.Current().Val
		//		{
		//			k, v :=...
		//		}
		// }
		body := X.Block(kv, n.Body)
		forStmt = X.ForStmt(nil, cond, nil, body)
	} else {
		body := X.Block1(kv, n.Body.List...)
		forStmt = X.ForStmt(nil, cond, nil, body)
	}
	return
}
