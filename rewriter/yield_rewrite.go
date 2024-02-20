package rewriter

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"github.com/goghcrow/go-loader"
	"github.com/goghcrow/go-matcher"
	"github.com/goghcrow/go-matcher/combinator"
	"golang.org/x/tools/go/ast/astutil"
)

type yieldRewriter struct {
	rewriter *rewriter
	pkg      loader.Pkg

	funcTyp  *ast.FuncType
	funcBody *ast.BlockStmt
	*yieldAst

	// file scope cache
	rewriteRetCache map[ast.Node]bool

	symCnt int // for unique symbol
}

func mkYieldRewriter(r *rewriter, pkg loader.Pkg) func(*astutil.Cursor, loader.Pkg) bool {
	return (&yieldRewriter{
		rewriter:        r,
		pkg:             pkg,
		rewriteRetCache: map[ast.Node]bool{},
	}).rewrite
}

func (r *yieldRewriter) rewrite(c *astutil.Cursor, pkg loader.Pkg) bool {
	switch f := c.Node().(type) {
	case *ast.FuncDecl:
		if r.rewriter.isYieldFuncDecl(f) {
			r.rewriteYieldFunc(f.Type, f.Body)
			c.Replace(f)
		}
		return true
	case *ast.FuncLit:
		if r.rewriter.isYieldFuncLit(f) {
			r.rewriteYieldFunc(f.Type, f.Body)
			c.Replace(f)
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
		r.rewriter.yieldFuncRetParamTy(r.pkg, funTy),
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
	// pass0
	// 1. rewrite `return` or `return nil` to return seq.Return() in yield func
	//
	// in the old implementation, we replace the trival block body of
	// "ifStmt / forStmt / blockStmt..." to the rewritten version (pass2)
	// for keeping the returnStmt (return => return seq.Return[T]()) in rewriteStmt.
	//
	// now, we rewrite returnStmt in yieldFunc (pass0) instead of in rewriteStmt (pass2),
	// so, the original node can be returned directly in trival branch of "if/for/block...".
	// the code is more intuitive.
	//
	// in the old returnNormalRequired func, "return Normal()" is added in all kindTrival blocks,
	// now, we only add for non-terminating blocks.
	//
	// and why rewriting ReturnStmt firstly?
	// cause of easy to recognize yield func before any rewriting
	//
	// 2. extract out define in for-init/switch-init in yield func
	//
	// keep shadow semantics by adding extra block,
	// prevent name conflict name in the scope after rewriteYield
	// e.g.,
	//
	//	i := 0; for i := 0; ; { ... }
	//	=>
	//	i := 0; i := 0; for ; ; { ... }
	//	=>
	//	i := 0; { i := 0; for ; ; { ... } }
	//
	//	for $init; ; { ... }
	//	=>
	//	{
	//		$init
	//		for ; ; { ... }
	//	}

	r.rewriteReturnAndForSwitchInitStmtInYieldFun(r.funcBody)

	// pass1
	r.rewriteRanges(r.funcBody)

	// bootstrap
	// start(delay(func() { kindDelay }))
	following := mkBlock(kindDelay /*callback func lit body*/)

	// pass2
	// skipping ReturnStmt, which rewritten in pass0
	r.rewriteStmts(r.funcBody.List, 0, following)

	// pass3
	// we will rewrite break/continue in the second pass
	// because we don't know whether to keep break/continue in trival context,
	// or to replace with co.Break() co.Continue() in monadic context
	r.rewriteBreakContinues(following.block)

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
		// no need to check kindDelay in returnNormalRequired
		callDelay := r.CallDelay(following.block)
		children.pushReturn(callDelay, kindYield /*notice: not kindDelay*/)
		return children

	case *ast.EmptyStmt:
		// ↓↓ trival branch ↓↓
		// skip and rewrite next stmt
		return children

	case *ast.ExprStmt:
		// rewrite yield call
		if call, ok := r.rewriter.isYieldCall(r.pkg, stmt); ok {
			r.checkYieldCall(call) // typeCheck

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
		// rewrite branch in pass3
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
		// &stmt.Init maybe ptr of typed nil
		return r.rewriteSwitchStmt(
			stmt, &stmt.Init, stmt.Tag, stmt.Body, &stmt.Switch, children,
		)

	case *ast.TypeSwitchStmt:
		// ↓↓ non-trival branch ↓↓
		trivalAssign := r.mustNoYield(stmt.Assign)
		r.assert(trivalAssign, stmt.Assign, "yield not allowed")
		return r.rewriteSwitchStmt(
			stmt, &stmt.Init, stmt.Assign, stmt.Body, &stmt.Switch, children,
		)

	case *ast.ForStmt:
		// ↓↓ non-trival branch ↓↓
		return r.rewriteForStmt(stmt, children)

	// rewritten in pass1
	// case *ast.RangeStmt:

	case *ast.SelectStmt, *ast.CommClause,
		*ast.LabeledStmt, *ast.CaseClause,
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
	if children.returnNormalRequired(r.isTerminating) {
		// markCombined manually, cause of no need to check
		// when normal return required
		children.markCombined()
		// return Normal[T]()
		children.pushReturn(r.callNormal, kindNormal)
	}
}
func (r *yieldRewriter) checkYieldCall(call *ast.CallExpr) {
	v := r.pkg.TypeOf(call.Args[0])
	t := r.pkg.TypeOf(r.yieldAst.funRetParamTy)
	// generated codes have attached the type
	assert(v != nil && t != nil)

	arg := r.pkg.ShowNode(call.Args[0])
	r.assert(types.AssignableTo(v, t), call.Lparen,
		"yield(%s):"+
			" type mismatch, typeof(%s) is %s, "+
			"not assignable to return type %s",
		arg, arg,
		v.String(), t.String())
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
	stmt ast.Stmt, // switch | typeSwitch
	init *ast.Stmt,
	x ast.Node, // Tag of SwitchStmt | Assign of TypeSwitchStmt
	body *ast.BlockStmt, // maybe modified
	pos *token.Pos, // maybe modified
	children *block,
) *block {
	allCaseTrival := true
	var cases []ast.Stmt
	for _, it := range body.List {
		// yield is not supported in case expr, but
		// yield has no return, no need to assert
		clause := it.(*ast.CaseClause)
		caseBody := r.rewriteBlockStmt(X.Block(clause.Body...), kindSwitch)
		cases = append(cases, X.Case(clause.List, caseBody.block.List))
		allCaseTrival = allCaseTrival && caseBody.mustNoYield()
	}

	// trival routine
	trivalInit := r.mustNoYield(*init)
	if trivalInit && allCaseTrival {
		children.push(stmt, kindTrival)
		return children
	}

	// extract out init stmt if present
	if *init != nil {
		// details referring to comment in rewriteInitStmt
		assert(!isDefineStmt(*init))
		children = r.rewriteStmt(*init, false, children)
		r.assert(children != nil, stmt, "illegal state")
		*init = nil
		*pos = token.NoPos
	}

	// stmt.Init non trival
	if allCaseTrival {
		switchStmt := X.Switch(
			nil, // extracted out
			x,
			body,
		)
		children.push(switchStmt, kindTrival)
		return children
	}

	// not all case trival
	switchStmt := X.Switch(
		nil,
		x,
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

	if body.combineRequired() {
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
	if !children.combineRequired() {
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

func (r *yieldRewriter) rewriteReturnAndForSwitchInitStmtInYieldFun(body *ast.BlockStmt) {
	var (
		yieldFunStack = mkStack[bool](true) // default in yield func
		enter         = yieldFunStack.push
		exit          = yieldFunStack.pop
		inYieldFunc   = yieldFunStack.top

		isRetNil = func(ret *ast.ReturnStmt) bool {
			if ret.Results == nil {
				return true
			}
			if len(ret.Results) != 1 {
				return false
			}
			tyNil := types.Universe.Lookup("nil")
			return types.Identical(tyNil.Type(), r.pkg.TypeOf(ret.Results[0]))
		}
	)

	astutil.Apply(body, func(c *astutil.Cursor) bool {
		switch f := c.Node().(type) {
		case *ast.FuncDecl:
			if r.rewriteRetCache[f] {
				return false
			}
			r.rewriteRetCache[f] = true
			enter(r.rewriter.isYieldFuncDecl(f))
		case *ast.FuncLit:
			if r.rewriteRetCache[f] {
				return false
			}
			r.rewriteRetCache[f] = true
			enter(r.rewriter.isYieldFuncLit(f))
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
					log.Println("ignore return: " + r.pkg.ShowNode(n))
					assert(len(n.Results) == 1)
					c.InsertBefore(X.IgnoreExpr(n.Results[0]))
				}
				c.Replace(X.Return(r.CallReturn()))
			}

		case *ast.ForStmt:
			if inYieldFunc() && isDefineStmt(n.Init) {
				init := n.Init
				n.Init = nil
				n.For = token.NoPos
				c.Replace(X.Block(init, n))
			}
		case *ast.SwitchStmt:
			if inYieldFunc() && isDefineStmt(n.Init) {
				init := n.Init
				n.Init = nil
				n.Switch = token.NoPos
				c.Replace(X.Block(init, n))
			}
		case *ast.TypeSwitchStmt:
			if inYieldFunc() && isDefineStmt(n.Init) {
				init := n.Init
				n.Init = nil
				n.Switch = token.NoPos
				c.Replace(X.Block(init, n))
			}

		}
		return true
	})
}

//	for {
//			if true {
//				continue
//			} else {
//				Yield(0)
//			}
//		}
//
// WOULD BE REWRITTEN TO
//
//	Loop[int](Delay[int](func() Seq[int] {
//		if true {
//			return Continue[int]()
//		} else {
//			return Bind[int](0, Normal[int])
//		}
//		return Normal[int]() // notice here: redundant return
//	}))
//
// because, when terminating checking in pass2, code as follows
// `if true { continue } else { return Bind(...) }`
// but, continue will be rewritten to return Continue() in pass3,
// so, extra terminating checking is required to keep code clean
func (r *yieldRewriter) rewriteBreakContinues(body *ast.BlockStmt) {
	var (
		loopStack = mkStack[bool](false) // default in trival for
		enterLoop = loopStack.push
		exitLoop  = loopStack.pop
		inLoop    = loopStack.top

		switchStack = mkStack[bool](false) // default in trival for
		enterSwitch = switchStack.push
		exitSwitch  = switchStack.pop
		inSwitch    = switchStack.top

		funcLitStack = mkStack[*ast.FuncLit](nil)
		enterFuncLit = funcLitStack.push
		exitFuncLit  = funcLitStack.pop

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

		rmRedundantReturn = func(body *ast.BlockStmt) {
			if len(body.List) == 0 {
				return
			}
			last := body.List[len(body.List)-1]
			if ret, ok := last.(*ast.ReturnStmt); ok {
				isRetNormal := len(ret.Results) == 1 && ret.Results[0] == r.callNormal
				if isRetNormal {
					stmts := body.List[:len(body.List)-1]
					if r.isTerminating(X.Block(stmts...)) {
						body.List = stmts
					}
				}
			}
		}
	)

	astutil.Apply(body, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch n := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			enterLoop(true)
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			enterSwitch(true)
		case *ast.FuncLit:
			enterLoop(false)
			enterSwitch(false)
			enterFuncLit(n)
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
			exitFuncLit()
		case *ast.BranchStmt:
			n1 := doRewrite(n)
			if n1 != nil {
				c.Replace(n1)

				top := funcLitStack.top()
				if top == nil { // init-value,  in outer yield func
					rmRedundantReturn(body)
				} else { // in funcLit
					rmRedundantReturn(top.Body)
				}
			}
		}
		return true
	})
}

// try to rewrite and test whether containing yield
func (r *yieldRewriter) mustNoYield(stmt ast.Stmt) bool {
	if isNil(stmt) {
		return true
	}
	return !r.rewriter.containsYield(r.pkg, X.Block(stmt))

	// // isLast=false and kind=kindYield
	// // don't affect the result
	// b := mkBlock(kindYield)
	// r.rewriteStmt(stmt, false, b)
	// return b.mustNoYield()
}

func (r *yieldRewriter) isTerminating(s ast.Stmt) bool {
	m := r.rewriter.m.Matcher
	panicCallSites := make(map[*ast.CallExpr]bool)
	m.Match(r.pkg.Package, combinator.BuiltinCallee(m, "panic"), s,
		func(c *matcher.Cursor, ctx *matcher.MatchCtx) {
			panicCallSites[c.Node().(*ast.CallExpr)] = true
		},
	)

	return mkTerminationChecker(panicCallSites).isTerminating(s)
}

func (r *yieldRewriter) assert(
	ok bool,
	pos any,
	format string,
	a ...any,
) {
	r.rewriter.assert(r.pkg, ok, pos, format, a...)
}
