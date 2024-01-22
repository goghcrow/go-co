package rewriter

import (
	"go/ast"
)

type yieldAst struct {
	seqImportedName string
	funRetParamTy   ast.Expr // generator element type
}

func mkYieldAst(seqName string, retParamTy ast.Expr) *yieldAst {
	return &yieldAst{
		seqImportedName: seqName,
		funRetParamTy:   retParamTy,
	}
}

func (y *yieldAst) SeqSelect(name string) ast.Expr {
	return X.PkgSelect(y.seqImportedName, name)
}

func (y *yieldAst) SeqIndex(name string) *ast.IndexExpr {
	return X.Index(
		y.SeqSelect(name),
		y.funRetParamTy,
	)
}

func (y *yieldAst) SeqFun(name string) *ast.IndexExpr {
	return y.SeqIndex(name)
}

func (y *yieldAst) SeqType(name string) *ast.IndexExpr {
	return y.SeqIndex(name)
}

func (y *yieldAst) SeqCall(name string, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  y.SeqFun(name),
		Args: args,
	}
}

func (y *yieldAst) Thunk(body *ast.BlockStmt) *ast.FuncLit {
	return &ast.FuncLit{
		Type: &ast.FuncType{
			Params: X.Fields(),
			Results: X.Fields(
				X.TypeField(y.SeqType(cstSeq)),
			),
		},
		Body: body,
	}
}

func (y *yieldAst) CallStart(body *ast.BlockStmt) *ast.CallExpr {
	return y.SeqCall(cstStart,
		y.CallDelay(body),
	)
}

func (y *yieldAst) CallNormal() *ast.CallExpr {
	return y.SeqCall(cstNormal)
}

func (y *yieldAst) CallReturn() *ast.CallExpr {
	return y.SeqCall(cstReturn)
}

func (y *yieldAst) CallBreak() *ast.CallExpr {
	return y.SeqCall(cstBreak)
}

func (y *yieldAst) CallContinue() *ast.CallExpr {
	return y.SeqCall(cstContinue)
}

func (y *yieldAst) CallDelay(body *ast.BlockStmt) *ast.CallExpr {
	return y.SeqCall(
		cstDelay,
		y.Thunk(body),
	)
}

func (y *yieldAst) CallBind(v ast.Expr, body *ast.BlockStmt) *ast.CallExpr {
	return y.SeqCall(cstBind,
		v,
		y.Thunk(body),
	)
}

func (y *yieldAst) CallCombine(s1, s2 *ast.BlockStmt) *ast.CallExpr {
	return y.SeqCall(cstCombine,
		y.CallDelay(s1),
		y.CallDelay(s2),
	)
}

func (y *yieldAst) CallFor(cond, post, body ast.Expr) *ast.CallExpr {
	if isNil(cond) && isNil(post) {
		return y.SeqCall(cstLoop, body)
	}
	if isNil(post) {
		return y.SeqCall(cstWhile, cond, body)
	}
	return y.SeqCall(cstFor, cond, post, body)
}

func (y *yieldAst) ForCondFun(cond ast.Expr) *ast.FuncLit {
	if isNil(cond) {
		return nil
	}
	return &ast.FuncLit{
		Type: &ast.FuncType{
			Params: X.Fields(),
			Results: X.Fields(
				X.TypeField(X.Ident("bool")),
			),
		},
		Body: X.Block(X.Return(cond)),
	}
}

func (y *yieldAst) ForPostFun(post ast.Stmt) *ast.FuncLit {
	if isNil(post) {
		return nil
	}
	return &ast.FuncLit{
		Type: &ast.FuncType{
			Params:  X.Fields(),
			Results: X.Fields(),
		},
		Body: X.Block(post),
	}
}
