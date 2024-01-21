package rewriter

import "go/ast"

type stmtKind int

const (
	kindTrival stmtKind = iota // native golang ast
	kindDelay                  // mark block which constructing seq.Delay(func() { $block$ })
	kindIf                     // if stmt which containing yield
	kindSwitch                 // switch stmt which calling yield

	// ↓↓↓ MUST be constructed with return stmt
	kindNormal  // stmt which calling seq.Normal()
	kindYield   // stmt which calling seq.Bind()
	kindCombine // stmt which calling seq.Combine()
	kindFor     // stmt which calling seq.For()
	// ↑↑↑ MUST be constructed with return stmt
)

type block struct {
	block *ast.BlockStmt
	kinds []stmtKind // kind of stmt in block.List

	// mark the block which stmt belongs to
	// e.g., seq.Delay, seq.For, seq.Bind, ...
	// e.g., kindDelay/ kindFor/ kindYield ...
	kind stmtKind

	frozen         bool // whether return stmt appended
	combineChecked bool
}

func mkBlock(kind stmtKind) *block {
	return &block{
		block:          X.Block(),
		kind:           kind,
		combineChecked: true,
	}
}

func (b *block) len() int {
	return len(b.block.List)
}

// marking combine checking done
func (b *block) markCombined() {
	b.combineChecked = true
}

// refactoring: auto combineIfNecessary and return new *block?
func (b *block) push(stmt ast.Stmt, kind stmtKind) {
	// should call rewriter.combineIfNecessary after push calling
	assert(b.combineChecked)
	assert(!b.frozen)
	b.block.List = append(b.block.List, stmt)
	b.kinds = append(b.kinds, kind)
	b.combineChecked = false
}

func (b *block) pushReturn(call *ast.CallExpr, whatToCall stmtKind) {
	assert(whatToCall >= kindNormal)
	stmt := X.Return(call)
	b.push(stmt, whatToCall)
	b.frozen = true
}

func (b *block) pop() (ast.Stmt, stmtKind) {
	assert(b.len() > 0)
	idx := len(b.block.List) - 1
	stmt := b.block.List[idx]
	kind := b.kinds[idx]

	b.block.List = b.block.List[:idx]
	b.kinds = b.kinds[:idx]
	b.frozen = false
	return stmt, kind
}

func (b *block) lastKind() stmtKind {
	assert(b.len() > 0)
	return b.kinds[len(b.kinds)-1]
}

func (b *block) lastStmt() ast.Stmt {
	assert(b.len() > 0)
	return b.block.List[len(b.block.List)-1]
}

func (b *block) last() (ast.Stmt, stmtKind) {
	assert(b.len() > 0)
	idx := len(b.block.List) - 1
	return b.block.List[idx], b.kinds[idx]
}

func (b *block) mayContainsYield() bool {
	//	kindNormal: no yield
	// 	kindTrival: no yield
	//  kindReturn: no yield
	//	kindYield: just yield
	//	kindIf: if-stmt marked kindIf when body contains yield, otherwise trival
	//	kindFor: for-stmt marked kindFor when body/post contains yield, otherwise trival
	//	kindCombine: non-trival combined, may contain yield
	//	kindSwitch: the same as kindIf
	//	kindDelay: shouldn't exist, details can refer to comment in BlockStmt rewritten

	// fast routine
	if b.len() == 0 {
		return false
	}
	switch b.lastKind() {
	case kindYield, kindFor, kindCombine: // ending with return
		return true
	default: // make ide happy
	}

	// mustn't test the last stmt only when if/switch/delay
	// should trace back to find the first one containing yield
	// e.g.,
	//	if {
	//		Yield(1)
	//	}
	//  println(42)
	// =>
	// if {
	// 		return Bind(1, func() Seq[T] {...
	// }
	// println(42)
	for i := len(b.kinds) - 1; i >= 0; i-- {
		switch b.kinds[i] {
		case kindIf, kindSwitch:
			return true
		case kindDelay:
			// details can refer to comment in BlockStmt rewritten
			panic("shouldn't be here")
		default: // make ide happy
		}
	}

	return false
}

func (b *block) needCombine() bool {
	return b.len() > 0 && b.lastKind() != kindTrival
}

func (b *block) mustNoYield() bool {
	return !b.mayContainsYield()
}

func (b *block) requireReturnNormal() bool {
	assert(b.kind == kindDelay ||
		b.kind == kindFor || b.kind == kindIf)

	if b.len() == 0 {
		// e.g., following of last yield stmt has empty block
		// so return-normal needed
		return true
	}

	// notice:
	// even though for-block is empty, `return Normal()` MUSTN'T be appended
	// e.g., for { } => for { return Normal() }
	// if return appended, control flow will be broken

	last, kind := b.last()
	switch kind {
	default:
		return false
		// return-normal has appended to all other non-trival stmts manually
		// 	kindNormal: children.pushReturn(callNormal, kindNormal)
		// 	kindReturn: children.pushReturn(callReturn, kindReturn)
		// 	kindYield: children.pushReturn(callBind, kindYield)
		// 	kindCombine: children.pushReturn(callCombine, kindCombine)
		// 	kindFor: children.pushReturn(callFor, kindFor)
		// 	kindIf, kindSwitch, kindTrival: ↓↓↓
	case kindIf, kindSwitch:
		return !allBranchesEndWithReturn(last)
	case kindTrival:
		// return-nil has rewritten to kindTrival return-stmt in pass0
		return !instanceof[*ast.ReturnStmt](last)
	}
}
