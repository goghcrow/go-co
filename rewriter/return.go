package rewriter

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// modified from go/src/go/types/return.go
// cropping label control flow, cause of no supporting in yield func

type terminationChecker struct {
	panicCallSites map[*ast.CallExpr]bool
}

func mkTerminationChecker(panicCallSites map[*ast.CallExpr]bool) *terminationChecker {
	return &terminationChecker{
		panicCallSites: panicCallSites,
	}
}

func (check *terminationChecker) isTerminating(s ast.Stmt) bool {
	switch s := s.(type) {
	default:
		panic("unreachable")

	case *ast.BadStmt, *ast.DeclStmt, *ast.EmptyStmt, *ast.SendStmt,
		*ast.IncDecStmt, *ast.AssignStmt, *ast.GoStmt, *ast.DeferStmt,
		*ast.RangeStmt /*range empty*/ :
		// no chance

	case *ast.LabeledStmt:
		return check.isTerminating(s.Stmt)

	case *ast.ExprStmt:
		// calling the predeclared (possibly parenthesized) panic() function is terminating
		if call, ok := astutil.Unparen(s.X).(*ast.CallExpr); ok && check.panicCallSites[call] {
			return true
		}

	case *ast.ReturnStmt:
		return true

	case *ast.BranchStmt:
		if s.Tok == token.GOTO || s.Tok == token.FALLTHROUGH {
			return true
		}

	case *ast.BlockStmt:
		return check.isTerminatingList(s.List)

	case *ast.IfStmt:
		if s.Else != nil &&
			check.isTerminating(s.Body) &&
			check.isTerminating(s.Else) {
			return true
		}

	case *ast.SwitchStmt:
		return check.isTerminatingSwitch(s.Body)

	case *ast.TypeSwitchStmt:
		return check.isTerminatingSwitch(s.Body)

	case *ast.SelectStmt:
		for _, s := range s.Body.List {
			cc := s.(*ast.CommClause)
			if !check.isTerminatingList(cc.Body) || hasBreakList(cc.Body) {
				return false
			}

		}
		return true

	case *ast.ForStmt:
		if s.Cond == nil && !hasBreak(s.Body) {
			return true
		}
	}

	return false
}

func (check *terminationChecker) isTerminatingList(list []ast.Stmt) bool {
	// trailing empty statements are permitted - skip them
	for i := len(list) - 1; i >= 0; i-- {
		if _, ok := list[i].(*ast.EmptyStmt); !ok {
			return check.isTerminating(list[i])
		}
	}
	return false // all statements are empty
}

func (check *terminationChecker) isTerminatingSwitch(body *ast.BlockStmt) bool {
	hasDefault := false
	for _, s := range body.List {
		cc := s.(*ast.CaseClause)
		if cc.List == nil {
			hasDefault = true
		}
		if !check.isTerminatingList(cc.Body) || hasBreakList(cc.Body) {
			return false
		}
	}
	return hasDefault
}

func hasBreak(s ast.Stmt) bool {
	switch s := s.(type) {
	default:
		panic("unreachable")

	case *ast.BadStmt, *ast.DeclStmt, *ast.EmptyStmt, *ast.ExprStmt,
		*ast.SendStmt, *ast.IncDecStmt, *ast.AssignStmt, *ast.GoStmt,
		*ast.DeferStmt, *ast.ReturnStmt,

		*ast.SwitchStmt, *ast.TypeSwitchStmt,
		*ast.SelectStmt, *ast.ForStmt, *ast.RangeStmt:
		// no chance

	case *ast.LabeledStmt:
		return hasBreak(s.Stmt)

	case *ast.BranchStmt:
		if s.Tok == token.BREAK {
			if s.Label == nil {
				panic("labelled break not supported")
			}
			return true
		}

	case *ast.BlockStmt:
		return hasBreakList(s.List)

	case *ast.IfStmt:
		if hasBreak(s.Body) ||
			s.Else != nil && hasBreak(s.Else) {
			return true
		}

	case *ast.CaseClause:
		return hasBreakList(s.Body)

	case *ast.CommClause:
		return hasBreakList(s.Body)

	}

	return false
}

func hasBreakList(list []ast.Stmt) bool {
	for _, s := range list {
		if hasBreak(s) {
			return true
		}
	}
	return false
}
