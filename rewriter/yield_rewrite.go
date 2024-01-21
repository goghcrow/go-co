package rewriter

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

type yieldRewriter struct {
	rewriter *rewriter

	funcTyp  *ast.FuncType
	funcBody *ast.BlockStmt
	*yieldAst

	// file scope cache
	rewriteRetCache map[ast.Node]bool
}

func mkYieldRewriter(r *rewriter) *yieldRewriter {
	return &yieldRewriter{
		rewriter:        r,
		rewriteRetCache: map[ast.Node]bool{},
	}
}

func (r *yieldRewriter) rewrite(c *astutil.Cursor) bool {
	typeof := r.rewriter.TypeOf

	switch n := c.Node().(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			return true
		}
		if r.rewriter.isYieldFunc(typeof(n.Name)) {
			r.rewriteYieldFunc(n.Type, n.Body)
			c.Replace(n)
		}
		return true
	case *ast.FuncLit:
		if n.Body == nil {
			return true
		}
		if r.rewriter.isYieldFunc(typeof(n)) {
			r.rewriteYieldFunc(n.Type, n.Body)
			c.Replace(n)
		}
		return true
	}
	return true
}

func (r *yieldRewriter) rewriteYieldFunc(
	funTy *ast.FuncType,
	body *ast.BlockStmt,
) {
	r.funcTyp = funTy
	r.funcBody = body
	r.yieldAst = mkYieldAst(
		r.rewriter.seqImportedName,
		r.rewriter.yieldFuncRetParamTy(funTy),
	)
	// not support recursive, no need stack
	defer func() {
		r.funcTyp = nil
		r.funcBody = nil
		r.yieldAst = nil
	}()

	// may be ignored, rewriteIter do the same thing
	r.rewriteYieldFuncResult()
	r.rewriteYieldFuncBody()
}

func (r *yieldRewriter) rewriteYieldFuncResult() {
	r.funcTyp.Results.List[0].Type = r.SeqType(cstIterator)
}

// >>> return Start(Delay[T](func() Seq[T] { ... }))
func (r *yieldRewriter) rewriteYieldFuncBody() {
	// bootstrap
	// start(delay(func() { kindDelay }))
	following := mkBlock(kindDelay /*callback func lit body*/)

	// pass0
	// rewrite `return` or `return nil` to return seq.Return() in yield func
	// in the old implementation, we replace the trival block body of
	// "ifStmt / forStmt / blockStmt..." to the rewriting result
	// for keeping the rewiring result of returnStmt
	// (return => return seq.Return[T]()) in rewriteStmt.
	// now, we rewrite returnStmt in yieldFunc (pass0) instead of in rewriteStmt (pass1),
	// so, the original node can be returned directly in trival branch of "if/for/block...".
	// the code is more intuitive.
	// in the old requireReturnNormal func, "return Normal()" is added in all kindTrival blocks,
	// now, we only add for blocks not ending with returnStmt.
	// why rewriting return first?
	// cause of easy to recognize yield func before any rewriting
	r.rewriteReturn(r.funcBody)

	// pass1
	// skipping ReturnStmt, which rewritten in pass0
	r.rewriteStmts(r.funcBody.List, 0, following)

	// pass2
	// we will rewrite break/continue in the second pass
	// because we don't know whether to keep break/continue in trival context,
	// or to replace with co.Break() co.Continue() in monadic context
	r.rewriteBreakContinue(following.block)

	returnCallStart := X.Return(r.CallStart(following.block))
	r.funcBody.List = []ast.Stmt{returnCallStart}
}

func (r *yieldRewriter) rewriteStmts(
	stmts []ast.Stmt,
	idx int,
	children *block,
) {
	done := idx >= len(stmts)
	if done {
		if children.kind == kindDelay {
			r.generateLastNormalIfNecessary(children)
		}
		return
	}

	isLast := idx == len(stmts)-1
	following := r.rewriteStmt(stmts[idx], isLast, children)
	if following == nil {
		return
	}

	if isLast {
		if children.kind == kindDelay {
			r.generateLastNormalIfNecessary(children)
		}
	} else {
		following = r.combineIfNecessary(following)
		r.rewriteStmts(stmts, idx+1, following)
	}
}

func (r *yieldRewriter) rewriteStmt(
	stmt ast.Stmt,
	isLast bool, // generateLastNormalIfNecessary
	children *block,
) *block {
	switch stmt := stmt.(type) {

	case nil:
		panic("illegal state")

	case *ast.BadStmt:
		r.assert(false, stmt, "bad stmt")
		panic("make compiler happy")

	case *ast.BlockStmt:
		following := r.rewriteBlockStmt(stmt, kindDelay)
		if following.mustNoYield() {
			children.push(stmt, kindTrival)
			return children
		}

		// kindDelay are used for marking the body (block) of seq.Delay(func(){...})
		// which is a callback of seq.Start/Combine/Bind/For/While/Loop ...
		// we mark the blockStmt containing yield with kindYield instead of kindDelay
		// to keep the logical consistency,
		// so, all kindDelay shouldn't appear in block.List, and
		// no need to check kindDelay in requireReturnNormal
		callDelay := r.CallDelay(following.block)
		children.pushReturn(callDelay, kindYield /*notice: not kindDelay*/)
		return children

	case *ast.EmptyStmt:
		// ↓↓ trival branch ↓↓
		// skip and rewrite next stmt
		return children

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
				return nil                                 // last stmt, no following
			} else {
				// switch children to empty binding callback body (following)
				// no need combine cause of body is empty
				r.assert(following != children, stmt, "illegal state")
				return following
			}
		} else {
			// ↓↓ trival branch ↓↓
			// rewrite next stmt in current block
			// no need combine cause of prev stmt is trival
			children.push(stmt, kindTrival)
			return children
		}

	case *ast.BranchStmt:
		// ↓↓ trival branch ↓↓
		// rewrite branch in pass2
		switch stmt.Tok {
		case token.BREAK, token.CONTINUE, token.FALLTHROUGH:
			// fallthrough supported only in trival switch node
			children.push(stmt, kindTrival)
			// ignore dead code after return break/ continue,
			// only goto or labeled-stmt can reach the stmts after break/continue
			// no need to rewrite next, cause of goto or labeled-stmt unsupported
			return nil // ignore dead code, no following
		case token.GOTO:
			r.assert(false, stmt, "goto not supported")
			panic("make compiler happy")
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
			return nil // no following
		} else {
			return children
		}

	case *ast.SwitchStmt:
		// ↓↓ non-trival branch ↓↓
		return r.rewriteSwitchStmt(stmt, children)

	case *ast.ForStmt:
		// ↓↓ non-trival branch ↓↓
		return r.rewriteForStmt(stmt, children)

	case *ast.RangeStmt:
		r.assert(false, stmt, "implement me")
		panic("make compiler happy")

	case *ast.SelectStmt, *ast.CommClause,
		*ast.LabeledStmt, *ast.CaseClause,
		*ast.TypeSwitchStmt,
		*ast.DeferStmt:
		r.assert(false, stmt, "%T implement me", stmt)
		panic("make compiler happy")

	// rewritten in pass0
	// case *ast.ReturnStmt:
	// // ↓↓ non-trival branch ↓↓
	// // Notice: ignore stmt.Results
	// r.generateReturnStmt(children)
	// return

	default:
		// ↓↓ trival branch ↓↓
		// all other stmt are trival,
		// no rewriting, no combine
		children.push(stmt, kindTrival)
		return children
	}
}

func (r *yieldRewriter) rewriteBlockStmt(
	body *ast.BlockStmt,
	kind stmtKind,
) *block {
	following := mkBlock(kind)
	r.rewriteStmts(body.List, 0, following)
	return following
}

// additional return-normal() required
// when last stmt is kindIf / kindSwitch / kindTrival
func (r *yieldRewriter) generateLastNormalIfNecessary(children *block) {
	if children.requireReturnNormal(r.rewriter) {
		// markCombined manually, cause of no need to check
		// when normal return required
		children.markCombined()
		// return Normal[T]()
		callNormal := r.CallNormal()
		children.pushReturn(callNormal, kindNormal)
	}
}

// return Bind($v, func() Seq[T] { $following })
func (r *yieldRewriter) rewriteYieldCall(
	call *ast.CallExpr,
	children *block,
) *block {
	// bind(v, func() { kindDelay })
	following := mkBlock(kindDelay /*callback func lit body*/)
	v := call.Args[0]
	callBind := r.CallBind(v, following.block)
	children.pushReturn(callBind, kindYield)
	return following
}

func (r *yieldRewriter) rewriteIfStmt(
	stmt *ast.IfStmt,
	children *block,
) {
	// merge elif stmt
	// 		if {} else { if {} ... }
	// => 	if {} else if {} ...
	unwrapIf := func(block *ast.BlockStmt) ast.Stmt {
		if block == nil {
			return nil
		}
		if len(block.List) == 1 {
			if iff, ok := block.List[0].(*ast.IfStmt); ok {
				return iff
			}
		}
		return block
	}

	switch alt := stmt.Else.(type) {
	case nil:
		body := r.rewriteBlockStmt(stmt.Body, kindIf)
		if body.mustNoYield() {
			children.push(stmt, kindTrival)
			return
		}

		elsStmt := unwrapIf(nil)
		iff := X.IfStmt(stmt.Init, stmt.Cond, body.block, elsStmt)
		children.push(iff, kindIf)
		return

	case *ast.BlockStmt:
		body := r.rewriteBlockStmt(stmt.Body, kindIf)
		els := r.rewriteBlockStmt(alt, kindIf)
		isTrival := body.mustNoYield() && els.mustNoYield()
		if isTrival {
			children.push(stmt, kindTrival)
			return
		}

		elsStmt := unwrapIf(els.block)
		iff := X.IfStmt(stmt.Init, stmt.Cond, body.block, elsStmt)
		children.push(iff, kindIf)
		return

	case *ast.IfStmt:
		body := r.rewriteBlockStmt(stmt.Body, kindIf)
		els := mkBlock(kindIf)
		r.rewriteIfStmt(alt, els)
		isTrival := body.mustNoYield() && els.mustNoYield()
		if isTrival {
			children.push(stmt, kindTrival)
			return
		}

		elsStmt := unwrapIf(els.block)
		iff := X.IfStmt(stmt.Init, stmt.Cond, body.block, elsStmt)
		children.push(iff, kindIf)
		return
	default:
		panic("unreached")
	}
}

func (r *yieldRewriter) rewriteSwitchStmt(
	stmt *ast.SwitchStmt,
	children *block,
) *block {
	trivalInit := r.mustNoYield(stmt.Init)
	allCaseTrival := true
	var cases []ast.Stmt
	for _, it := range stmt.Body.List {
		// yield is not supported in case expr, but
		// yield has no return, no need to assert
		clause := it.(*ast.CaseClause)
		caseBody := r.rewriteBlockStmt(X.Block(clause.Body...), kindSwitch)
		cases = append(cases, X.Case(clause.List, caseBody.block.List))
		allCaseTrival = allCaseTrival && caseBody.mustNoYield()
	}

	// trival routine
	if trivalInit && allCaseTrival {
		children.push(stmt, kindTrival)
		return children
	}

	// extract out init stmt if present
	if stmt.Init != nil {
		// details referring to comment in rewriteInitStmt
		assert(!isDefineStmt(stmt.Init))
		children = r.rewriteStmt(stmt.Init, false, children)
		r.assert(children != nil, stmt, "illegal state")
		stmt.Init = nil
		stmt.Switch = token.NoPos
	}

	// stmt.Init non trival
	if allCaseTrival {
		switchStmt := X.SwitchStmt(
			nil, // extracted out
			stmt.Tag,
			stmt.Body,
		)
		children.push(switchStmt, kindTrival)
		return children
	}

	// not all case trival
	switchStmt := X.SwitchStmt(
		nil,
		stmt.Tag,
		X.Block(cases...),
	)
	children = r.combineIfNecessary(children)
	children.push(switchStmt, kindSwitch)
	return children
}

func (r *yieldRewriter) rewriteForStmt(
	stmt *ast.ForStmt,
	children *block,
) *block {
	body := r.rewriteBlockStmt(stmt.Body, kindFor)

	trivalInit := r.mustNoYield(stmt.Init)
	trivalPost := r.mustNoYield(stmt.Post)
	trivalBody := body.mustNoYield()

	// trival routine
	allTrival := trivalBody && trivalInit && trivalPost
	if allTrival {
		children.push(stmt, kindTrival)
		return children
	}

	// extract out init stmt if present
	if stmt.Init != nil {
		// details referring to comment in rewriteInitStmt
		assert(!isDefineStmt(stmt.Init))
		children = r.rewriteStmt(stmt.Init, false, children)
		r.assert(children != nil, stmt, "illegal state")
		stmt.Init = nil
		stmt.For = token.NoPos
	}

	// trival routine
	if trivalBody && trivalPost {
		children = r.combineIfNecessary(children) // for init containing yield
		children.push(stmt, kindTrival)
		return children
	}

	if trivalPost {
		callFor := r.CallFor(
			r.ForCondFun(stmt.Cond),
			r.ForPostFun(stmt.Post),
			r.CallDelay(body.block),
		)
		children = r.combineIfNecessary(children)
		children.pushReturn(callFor, kindFor)
		return children
	}

	if body.needCombine() {
		// combine(delay(body), delay(post))
		// rewriting by seq.Combine avoiding control flow analysis (merging body & post)

		// instead of combine last-stmt of body and post,
		// we use two delay-callings to isolate the scope of body and post,
		// prevent post can access the variable declared in body

		// https://go.dev/ref/spec#SimpleStmt
		// SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl
		// only simple stmt allowed in for-post,
		// yield-call allowed, return not allowed,
		// so, no need to add return-normal
		postBlock := mkBlock(kindDelay)
		r.rewriteStmt(stmt.Post, true, postBlock)
		assert(instanceof[*ast.ReturnStmt](postBlock.lastStmt()))

		r.generateLastNormalIfNecessary(body)

		callCombine := r.CallCombine(body.block, postBlock.block)
		newBody := mkBlock(body.kind)
		newBody.pushReturn(callCombine, kindCombine)
		body = newBody
	} else {
		// can't declare variable in for-post, name conflict free
		assert(!isDefineStmt(stmt.Post))
		body.markCombined()
		r.rewriteStmt(stmt.Post, true, body)
	}

	callFor := r.CallFor(
		r.ForCondFun(stmt.Cond),
		nil,
		r.CallDelay(body.block),
	)
	children = r.combineIfNecessary(children)
	children.pushReturn(callFor, kindFor)
	return children
}

func (r *yieldRewriter) combineIfNecessary(children *block) *block {
	children.markCombined()
	if !children.needCombine() {
		return children
	}

	//	Combine[T](
	//		Delay(func() Seq[T] { ... }),
	//		Delay(func() Seq[T] { ... }),
	//	)
	// combine(delay(func() { $1 }), delay(func() { $2 }))
	// $1 need to make block marked kindDelay for testing normal return
	// $2 just mark kindDelay for following

	current := mkBlock(kindDelay /*callback fst delay body*/)
	current.push(children.pop())
	r.generateLastNormalIfNecessary(current)

	following := mkBlock(kindDelay /*callback snd delay body*/)
	callCombine := r.CallCombine(current.block, following.block)
	children.pushReturn(callCombine, kindCombine)

	return following
}

func (r *yieldRewriter) rewriteReturn(body *ast.BlockStmt) {
	var (
		yieldFunStack = mkStack[bool](true) // default in yield func
		enter         = yieldFunStack.push
		exit          = yieldFunStack.pop
		inYieldFunc   = yieldFunStack.top

		typeof   = r.rewriter.TypeOf
		isYield  = r.rewriter.isYieldFunc
		isRetNil = func(ret *ast.ReturnStmt) bool {
			if ret.Results == nil {
				return true
			}
			if len(ret.Results) != 1 {
				return false
			}
			tyNil := types.Universe.Lookup("nil")
			return types.Identical(tyNil.Type(), typeof(ret.Results[0]))
		}
	)

	astutil.Apply(body, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.FuncDecl:
			if r.rewriteRetCache[n] {
				return false
			}
			r.rewriteRetCache[n] = true
			enter(isYield(typeof(n.Name)))
		case *ast.FuncLit:
			if r.rewriteRetCache[n] {
				return false
			}
			r.rewriteRetCache[n] = true
			enter(isYield(typeof(n)))
		}
		return true
	}, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.FuncDecl, *ast.FuncLit:
			exit()
		case *ast.ReturnStmt:
			if n.Return == token.NoPos {
				return true // skip generated node
			}
			// notice: only rewrite `return` or `return nil` stmt
			if inYieldFunc() {
				if !isRetNil(n) {
					log.Println("ignore return: " + r.rewriter.ShowNodeWithPos(n))
					assert(len(n.Results) == 1)
					c.InsertBefore(X.IgnoreExpr(n.Results[0]))
				}
				c.Replace(X.Return(r.CallReturn()))
			}
		}
		return true
	})
}

func (r *yieldRewriter) rewriteBreakContinue(body *ast.BlockStmt) {
	var (
		loopStack = mkStack[bool](false) // default in trival for
		enterLoop = loopStack.push
		exitLoop  = loopStack.pop
		inLoop    = loopStack.top

		switchStack = mkStack[bool](false) // default in trival for
		enterSwitch = switchStack.push
		exitSwitch  = switchStack.pop
		inSwitch    = switchStack.top

		doRewrite = func(n *ast.BranchStmt) (_ ast.Node) {
			switch n.Tok {
			case token.BREAK:
				if inLoop() || inSwitch() {
					return
				}
				r.assert(n.Label == nil, n, "break with label not supported")
				return X.Return(r.CallBreak())
			case token.CONTINUE:
				if inLoop() {
					return
				}
				r.assert(n.Label == nil, n, "continue with label not supported")
				return X.Return(r.CallContinue())
			case token.GOTO:
				r.assert(false, n, "goto not supported")
			case token.FALLTHROUGH:
				if inSwitch() {
					return
				}
				r.assert(false, n, "fallthrough not supported")
			default:
				panic("unreached")
			}
			return
		}
	)

	astutil.Apply(body, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			enterLoop(true)
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			enterSwitch(true)
		case *ast.FuncLit:
			enterLoop(false)
			enterSwitch(false)
		}
		return true
	}, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch n := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			exitLoop()
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			exitSwitch()
		case *ast.FuncLit:
			exitLoop()
			exitSwitch()
		case *ast.BranchStmt:
			n1 := doRewrite(n)
			if n1 != nil {
				c.Replace(n1)
			}
		}
		return true
	})
}

// try to rewrite and test whether containing yield
func (r *yieldRewriter) mustNoYield(stmt ast.Stmt) bool {
	if isNilNode(stmt) {
		return true
	}
	// isLast=false and kind=kindYield
	// don't affect the result
	b := mkBlock(kindYield)
	r.rewriteStmt(stmt, false, b)
	return b.mustNoYield()
}

func (r *rewriter) isTerminating(s ast.Stmt) bool {
	checker := mkTerminationChecker(r)
	checker.collectPanic(s)
	return checker.isTerminating(s)
}

func (r *yieldRewriter) assert(
	ok bool,
	pos ast.Node,
	format string,
	a ...any,
) {
	r.rewriter.assert(ok, pos, format, a...)
}