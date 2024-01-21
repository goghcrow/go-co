package rewriter

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/goghcrow/go-ast-matcher"
	"golang.org/x/tools/go/types/typeutil"
)

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Ast Factory ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

var X = factor{}

type factor struct{} // ast factory

func (factor) Ident(name string) *ast.Ident {
	return &ast.Ident{
		Name: name,
	}
}

func (factor) Index(x, idx ast.Expr) *ast.IndexExpr {
	return &ast.IndexExpr{
		X:     x,
		Index: idx,
	}
}

func (factor) Select(x ast.Expr, sel string) ast.Expr {
	return &ast.SelectorExpr{
		X:   x,
		Sel: X.Ident(sel),
	}
}

func (factor) PkgSelect(pkgName, name string) ast.Expr {
	assert(pkgName != "_")
	if pkgName == "." {
		return X.Ident(name)
	}
	pkg := X.Ident(pkgName)
	return X.Select(pkg, name)
}

func (factor) TypeField(typ ast.Expr) *ast.Field {
	return &ast.Field{Type: typ}
}

func (factor) Fields(xs ...*ast.Field) *ast.FieldList {
	return &ast.FieldList{
		List: xs,
	}
}

func (factor) Call(fun ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fun,
		Args: args,
	}
}

func (factor) Assign(tok token.Token, lhs, rhs ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{lhs},
		Tok: tok,
		Rhs: []ast.Expr{rhs},
	}
}

func (factor) Define(lhs, rhs ast.Expr) *ast.AssignStmt {
	return X.Assign(token.DEFINE, lhs, rhs)
}

func (factor) IgnoreExpr(expr ast.Expr) *ast.AssignStmt {
	return X.Assign(token.ASSIGN, X.Ident("_"), expr)
}

func (factor) Return(xs ...ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: xs,
	}
}

func (factor) IfStmt(
	init ast.Stmt,
	cond ast.Expr,
	body *ast.BlockStmt,
	els ast.Stmt,
) *ast.IfStmt {
	return &ast.IfStmt{
		Init: init,
		Cond: cond,
		Body: body,
		Else: els,
	}
}

func (factor) Case(list []ast.Expr, body []ast.Stmt) *ast.CaseClause {
	return &ast.CaseClause{
		List: list,
		Body: body,
	}
}

func (factor) SwitchStmt(
	init ast.Stmt,
	tag ast.Expr,
	body *ast.BlockStmt,
) *ast.SwitchStmt {
	return &ast.SwitchStmt{
		Init: init,
		Tag:  tag,
		Body: body,
	}
}

func (factor) ForStmt(
	init ast.Stmt,
	cond ast.Expr,
	post ast.Stmt,
	body *ast.BlockStmt,
) *ast.ForStmt {
	return &ast.ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}
}

func (factor) Block(xs ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: xs,
	}
}

func (factor) Block1(x ast.Stmt, xs ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: append([]ast.Stmt{x}, xs...),
	}
}

func (factor) Stmt(n ast.Node) ast.Stmt {
	switch n := n.(type) {
	case ast.Expr:
		return &ast.ExprStmt{X: n}
	case ast.Stmt:
		return n
	default:
		panic("invalid")
	}
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Predication ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func isDefineStmt(stmt ast.Stmt) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	return ok && assign.Tok == token.DEFINE
}

func isCallStmtOf(info *types.Info, n ast.Node, callee types.Object) (*ast.CallExpr, bool) {
	expr, ok := n.(*ast.ExprStmt)
	if !ok {
		return nil, false
	}
	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return nil, false
	}
	return call, typeutil.Callee(info, call) == callee
}

func identicalWithoutTypeParam(x, y types.Type) bool {
	unwrapTyParam := func(ty types.Type) types.Type {
		if named, ok := ty.(*types.Named); ok {
			return named.Obj().Type()
		}
		return nil
	}
	return types.Identical(
		unwrapTyParam(x),
		unwrapTyParam(y),
	)
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓ stack for astutil.Apply ↓↓↓↓↓↓↓↓↓↓↓↓↓↓

type stack[T any] []T

func mkStack[T any](init T) *stack[T] {
	return &stack[T]{init}
}

func (s *stack[T]) push(v T) {
	*s = append(*s, v)
}
func (s *stack[T]) pop() T {
	v := s.top()
	*s = (*s)[:len(*s)-1]
	return v
}
func (s *stack[T]) top() T {
	return (*s)[len(*s)-1]
}
func (s *stack[T]) len() int {
	return len(*s)
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Others ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func isNilNode(n ast.Node) bool {
	return matcher.IsNilNode(n)
}

func instanceof[T any](x any) (ok bool) {
	_, ok = x.(T)
	return
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func assert(ok bool) {
	if !ok {
		panic("illegal state")
	}
}
