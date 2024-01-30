package rewriter

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/types/typeutil"
)

// https://stackoverflow.com/questions/14249217/how-do-i-know-im-running-within-go-test
var runningWithGoTest = flag.Lookup("test.v") != nil ||
	strings.HasSuffix(os.Args[0], ".test")

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

func (factor) Indices(x ast.Expr, indices ...ast.Expr) *ast.IndexListExpr {
	return &ast.IndexListExpr{
		X:       x,
		Indices: indices,
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

func (factor) Assign2(tok token.Token, lhs1, lhs2, rhs1, rhs2 ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{lhs1, lhs2},
		Tok: tok,
		Rhs: []ast.Expr{rhs1, rhs2},
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

func (factor) Switch(
	init ast.Stmt,
	x ast.Node,
	body *ast.BlockStmt,
) ast.Stmt {
	switch x := x.(type) {
	case ast.Expr:
		return X.SwitchStmt(init, x, body)
	case ast.Stmt:
		return X.TypeSwitchStmt(init, x, body)
	default:
		panic("invalid switch")
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

func (factor) TypeSwitchStmt(
	init ast.Stmt,
	assign ast.Stmt,
	body *ast.BlockStmt,
) *ast.TypeSwitchStmt {
	return &ast.TypeSwitchStmt{
		Init:   init,
		Assign: assign,
		Body:   body,
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
	if isNil(x) {
		return X.Block(xs...)
	} else {
		return &ast.BlockStmt{
			List: append([]ast.Stmt{x}, xs...),
		}
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

func (factor) Comment(pos token.Pos, text string) *ast.Comment {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "// " + line
	}
	text = strings.Join(lines, "\n")
	return &ast.Comment{
		Slash: pos,
		Text:  text,
	}
}

func (factor) Comments(xs ...*ast.Comment) *ast.CommentGroup {
	return &ast.CommentGroup{
		List: xs,
	}
}

func (factor) AppendComment(doc **ast.CommentGroup, comment *ast.Comment) {
	if *doc == nil {
		*doc = X.Comments()
	}
	(*doc).List = append((*doc).List, comment)
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Predication ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func isUnderline(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "_"
}

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

func isNil(n any) bool {
	if n == nil {
		return true
	}
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() {
		return true
	}
	return false
}

func instanceof[T any](x any) (ok bool) {
	_, ok = x.(T)
	return
}

func mkDir(dir string) string {
	dir, err := filepath.Abs(dir)
	panicIf(err)
	err = os.MkdirAll(dir, os.ModePerm)
	panicIf(err)
	return dir
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
