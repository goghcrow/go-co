package rewriter

import (
	"go/ast"
	"go/token"
)

func (r *yieldRewriter) rewriteStmts_(
	stmts []ast.Stmt,
	idx int,
	children *block,
) {
	isLast := idx == len(stmts)-1

	// recursive rewrite
	rewriteNext := func(following *block) {
		if isLast {
			if following.kind == kindDelay {
				r.generateLastNormalIfNecessary(following)
			}
			return
		}

		following = r.combineIfNecessary(following)
		r.rewriteStmts(stmts, idx+1, following)
	}

	// rewrite completely
	if idx >= len(stmts) {
		if children.kind == kindDelay {
			r.generateLastNormalIfNecessary(children)
		}
		return
	}

	// `children.push(stmt, kindTrival)`
	// MUST be appended to all trival (no rewriting) branches

	switch stmt := stmts[idx].(type) {

	case nil:
		panic("illegal state")

	case *ast.BadStmt:
		r.assert(false, stmt, "bad stmt")

	case *ast.BlockStmt:
		following := r.rewriteBlockStmt(stmt, kindDelay)
		if following.mustNoYield() {
			children.push(stmt, kindTrival)
			rewriteNext(children)
			return
		}

		// kindDelay are used for marking the body (block) of seq.Delay(func(){...})
		// which is a callback of seq.Start/Combine/Bind/For/While/Loop ...
		// we mark the blockStmt containing yield with kindYield instead of kindDelay
		// to keep the logical consistency,
		// so, all kindDelay shouldn't appear in block.List, and
		// no need to check kindDelay in requireReturnNormal
		callDelay := r.CallDelay(following.block)
		children.pushReturn(callDelay, kindYield /*notice: not kindDelay*/)
		rewriteNext(children)
		return

	case *ast.EmptyStmt:
		// ↓↓ trival branch ↓↓
		// skip and rewrite next stmt
		rewriteNext(children)

	case *ast.ExprStmt:
		// rewrite yield call
		if call, ok := r.rewriter.isYieldCall(stmt); ok {
			// ↓↓ non-trival branch ↓↓
			following := r.rewriteYieldCall(call, children)
			if isLast {
				// nothing after yield call
				// Bind($v, func() Seq[T] {
				// 		MAKE SURE NONEMPTY BODY
				// })
				r.generateLastNormalIfNecessary(following) // MUST
				// terminate the rewrite
			} else {
				// rewrite the next stmt in BIND BODY (children)
				// no need combine cause of body is empty
				r.assert(following != children, stmt, "illegal state")
				rewriteNext(following)
			}
		} else {
			// ↓↓ trival branch ↓↓
			// rewrite next stmt in current block
			// no need combine cause of prev stmt is trival
			children.push(stmt, kindTrival)
			rewriteNext(children)
		}

	case *ast.BranchStmt:
		// ↓↓ trival branch ↓↓
		switch stmt.Tok {
		case token.BREAK, token.CONTINUE, token.FALLTHROUGH:
			// fallthrough supported only in trival switch node
			children.push(stmt, kindTrival)
			// ignore dead code after return break/ continue,
			// only goto or labeled-stmt can reach the stmts after break/continue
			// no need to rewrite next, cause of goto or labeled-stmt unsupported
			// terminate the rewrite
		case token.GOTO:
			r.assert(false, stmt, "goto not supported")
		default:
			panic("unreached")
		}

	case *ast.IfStmt:
		// ↓↓ non-trival branch ↓↓
		// cause of no callback, no switching children
		r.rewriteIfStmt(stmt, children)
		if isLast {
			// MAKE SURE EVERY BRANCH END WITH RETURN STMT
			r.generateLastNormalIfNecessary(children)
			// terminate the rewrite
		} else {
			rewriteNext(children)
		}

	case *ast.ForStmt:
		// ↓↓ non-trival branch ↓↓
		following := r.rewriteForStmt(stmt, children)
		rewriteNext(following)

	case *ast.RangeStmt:
		r.assert(false, stmt, "implement me")

	case *ast.SelectStmt, *ast.CommClause,
		*ast.LabeledStmt, *ast.CaseClause,
		*ast.SwitchStmt, *ast.TypeSwitchStmt,
		*ast.DeferStmt:
		r.assert(false, stmt, "implement me")

	// rewritten in pass0
	// case *ast.ReturnStmt:
	// // ↓↓ non-trival branch ↓↓
	// // Notice: ignore stmt.Results
	// r.generateReturnStmt(children)
	// return

	default:
		// ↓↓ trival branch ↓↓
		// all other stmt are trival, no need to rewrite
		// and no need to combine
		children.push(stmt, kindTrival)
		rewriteNext(children)
	}
}

// >>> For[T](
// >>> 		func() bool { ... },
// >>> 		func() { ... },
// >>>		Delay[T](func() Seq[T] { ... }),
// >>> )
func (r *yieldRewriter) rewriteForStmt_(
	forStmt *ast.ForStmt,
	children *block,
) {
	body := r.rewriteBlockStmt(forStmt.Body, kindFor)
	trivalBody := body.mustNoYield()
	if trivalBody {
		children.push(forStmt, kindTrival)
		return // the last stmt is trival
	}

	// assert forStmt.Init mustNoYield
	// assert forStmt.Post mustNoYield

	if forStmt.Init != nil {
		children.push(forStmt.Init, kindTrival)
	}
	callFor := r.CallFor(
		r.ForCondFun(forStmt.Cond),
		r.ForPostFun(forStmt.Post),
		r.CallDelay(body.block),
	)
	children.pushReturn(callFor, kindFor)
	return // last stmt is not trival
}

// // return Return[T]()
// func (r *yieldRewriter) generateReturnStmt(children *block) {
// 	callReturn := r.CallReturn()
// 	children.pushReturn(callReturn, kindReturn)
// }
