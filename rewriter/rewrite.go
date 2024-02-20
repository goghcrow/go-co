package rewriter

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"github.com/goghcrow/go-ast-matcher"
	"github.com/goghcrow/go-imports"
	"github.com/goghcrow/go-loader"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/types/typeutil"
)

type rewriter struct {
	m astmatcher.ASTMatcher

	// global context
	iterType      types.Object
	yieldFunc     types.Object
	yieldFromFunc types.Object

	// file context
	coImportedName  string
	seqImportedName string
	yieldFuncDecls  map[*ast.FuncDecl]bool
	yieldFuncLits   map[*ast.FuncLit]bool
	comments        []*ast.CommentGroup
}

func mkRewriter(m astmatcher.ASTMatcher) *rewriter {
	return &rewriter{
		m:             m,
		iterType:      m.Loader.MustLookup(qualifiedIter),
		yieldFunc:     m.Loader.MustLookup(qualifiedYield),
		yieldFromFunc: m.Loader.MustLookup(qualifiedYieldFrom),
	}
}

// Yield is stmt, not expr
func (r *rewriter) isYieldCall(pkg loader.Pkg, n ast.Node) (*ast.CallExpr, bool) {
	// e.g. for Yield(1); not here; Yield(2) {  Yield(3) }
	return r.isCallStmtOf(pkg, n, r.yieldFunc)
}

// YieldFrom is stmt, not expr
func (r *rewriter) isYieldFromCall(pkg loader.Pkg, n ast.Node) (*ast.CallExpr, bool) {
	return r.isCallStmtOf(pkg, n, r.yieldFromFunc)
}

func (r *rewriter) isCallStmtOf(pkg loader.Pkg, n ast.Node, callee types.Object) (*ast.CallExpr, bool) {
	expr, ok := n.(*ast.ExprStmt)
	if !ok {
		return nil, false
	}
	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return nil, false
	}
	return call, pkg.Callee(call) == callee
}

func (r *rewriter) isYieldFuncDecl(f *ast.FuncDecl) bool {
	return r.yieldFuncDecls[f]
}

func (r *rewriter) isYieldFuncLit(f *ast.FuncLit) bool {
	return r.yieldFuncLits[f]
}

func (r *rewriter) yieldFuncRetParamTy(pkg loader.Pkg, f *ast.FuncType) ast.Expr {
	retTy := f.Results.List[0].Type
	idx, is := retTy.(*ast.IndexExpr)
	r.assert(pkg, is, f, "invalid yield func type")
	return idx.Index
}

func (r *rewriter) isIterator(ty types.Type) bool {
	return identicalWithoutTypeParam(r.iterType.Type(), ty)
}

func (r *rewriter) containsYield(pkg loader.Pkg, n *ast.BlockStmt) bool {

	return func() (contains bool) {
		var abort = new(int)
		defer func() {
			if r := recover(); r != nil && r != abort {
				panic(r)
			}
		}()
		astutil.Apply(n, func(c *astutil.Cursor) bool {
			switch n := c.Node().(type) {
			case *ast.FuncDecl, *ast.FuncLit:
				return false
			case *ast.CallExpr:
				callee := pkg.Callee(n)
				if callee == r.yieldFunc || callee == r.yieldFromFunc {
					contains = true
					panic(abort)
				}
			}
			return true
		}, nil)
		return
	}()
}

func (r *rewriter) assert(pkg loader.Pkg, ok bool, pos any, format string, a ...any) {
	if !ok {
		loc := "unknown"
		if !isNil(pos) {
			switch pos := pos.(type) {
			case ast.Node:
				loc = pkg.ShowNode(pos)
			case token.Pos:
				loc = pkg.Fset.Position(pos).String()
			case string:
				loc = pos
			}
		}
		panic(fmt.Sprintf(format, a...) + " in: " + loc)
	}
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewrite ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func (r *rewriter) rewriteAllFiles(printer FilePrinter) {
	coPkg := r.m.Loader.LookupPackage(pkgCoPath)
	if coPkg == nil {
		log.Printf("skip rewrite: no import %s\n", pkgCoPath)
		return
	}

	r.m.Loader.VisitAllFiles(func(f *loader.File) {
		if !imports.Uses(f, coPkg.Types) {
			log.Printf("skip file: %s\n", f.Filename)
			return
		}
		r.rewriteFile(f, printer)
	})
}

func (r *rewriter) rewriteFile(f *loader.File, printer FilePrinter) {
	parseOrImport := func(fset *token.FileSet, f *ast.File) (coName, seqName string) {
		coName = imports.ImportName(f, pkgCoPath, pkgCoName)
		assert(coName != "") // coPkg != nil
		seqName = imports.ImportName(f, pkgSeqPath, pkgSeqName)
		if seqName == "" {
			seqName = importSeqName
			astutil.AddNamedImport(fset, f, importSeqName, pkgSeqPath)
		}
		return
	}

	pkg := f.Package()

	// 1. init context
	r.coImportedName, r.seqImportedName = parseOrImport(f.Pkg.Fset, f.File) // parse import name
	r.comments = nil

	r.yieldFuncDecls = map[*ast.FuncDecl]bool{}
	r.yieldFuncLits = map[*ast.FuncLit]bool{}
	r.collectYieldFunc(pkg, f) // collect func with yield/yieldFrom call

	// 2. edit file
	log.Printf("visit file: %s\n", f.Filename)
	do := func(fn func(*astutil.Cursor, loader.Pkg) bool) {
		astutil.Apply(f.File, nil, func(c *astutil.Cursor) bool {
			return fn(c, pkg)
		})
	}

	// file level instance for file scope cache
	// notice: order matters
	do(r.attachComment)             // attach the original source to comments
	do(mkYieldFromRewriter(r, pkg)) // rewrite yieldFrom() to range yield() (range co.Iter)
	do(r.rewriteForRanges)          // rewrite range co.Iter to for loop co.Iter
	do(mkYieldRewriter(r, pkg))     // rewrite yield func
	do(r.rewriteIter)               // rewrite all co.Iter to seq.Iterator

	// 3. write file
	log.Printf("write file: %s\n", f.Filename)
	// clear free-floating comments, preventing confusing position of comments
	// https://github.com/golang/go/issues/20744
	f.File.Comments = r.comments
	printer(f.Filename, f)
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Collect YieldFunc ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// collect funs which containing yield / yieldFrom called directly by reference,
// make sure funcDecl | funcLit modified in place afterward
func (r *rewriter) collectYieldFunc(pkg loader.Pkg, f *loader.File) {
	var (
		yieldFunStack = mkStack[ast.Node /*FuncDecl|FuncLit*/](nil)
		enter         = yieldFunStack.push
		exit          = yieldFunStack.pop
		outer         = yieldFunStack.top
	)

	checkSignature := func(funTy types.Type, pos token.Pos) {
		msg := "invalid yield func signature, expect one co.Iter[T] return"

		sig, ok := funTy.(*types.Signature)
		r.assert(pkg, ok, pos, msg)
		rs := sig.Results()

		singleRet := rs != nil && rs.Len() == 1
		r.assert(pkg, singleRet, pos, msg)

		retTy := rs.At(0).Type()
		retIter := r.isIterator(retTy)
		r.assert(pkg, retIter, pos, msg)
	}

	info := f.Pkg.TypesInfo
	cache := map[ast.Node]bool{}
	astutil.Apply(f.File, func(c *astutil.Cursor) bool {
		switch f := c.Node().(type) {
		case *ast.FuncDecl:
			if cache[f] {
				return false
			}
			cache[f] = true
			enter(f)
		case *ast.FuncLit:
			if cache[f] {
				return false
			}
			cache[f] = true
			enter(f)
		}
		return true
	}, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.FuncDecl, *ast.FuncLit:
			exit()

		case *ast.CallExpr:
			callee := typeutil.Callee(info, n)
			if callee == r.yieldFunc || callee == r.yieldFromFunc {
				switch f := outer().(type) {
				case *ast.FuncDecl:
					checkSignature(info.TypeOf(f.Name), n.Pos())
					r.yieldFuncDecls[f] = true
				case *ast.FuncLit:
					checkSignature(info.TypeOf(f), n.Pos())
					r.yieldFuncLits[f] = true
				}
			}
		}
		return true
	})
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Attach comment ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func (r *rewriter) attachComment(c *astutil.Cursor, pkg loader.Pkg) bool {
	if runningWithGoTest {
		return true
	}
	switch f := c.Node().(type) {
	case *ast.FuncDecl:
		if r.isYieldFuncDecl(f) {
			// attach comment to func decl
			src := pkg.ShowNode(f)
			X.AppendComment(&f.Doc, X.Comment(f.Pos()-1, src))
			c.Replace(f)
		}
		return true
	case *ast.FuncLit:
		if r.isYieldFuncLit(f) {
			// attach comment to free-float
			src := pkg.ShowNode(f)
			r.comments = append(r.comments,
				X.Comments(X.Comment(f.Pos()-1, src)))
			c.Replace(f)
		}
		return true
	}
	return true
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewrite Range Generator ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func (r *rewriter) rewriteForRanges(c *astutil.Cursor, pkg loader.Pkg) bool {
	switch n := c.Node().(type) {
	case *ast.RangeStmt:
		if r.isIterator(pkg.TypeOf(n.X)) {
			c.Replace(r.rewriteForRange(pkg, n))
		}
		return true
	}
	return true
}

//	for x [:]= range $X {
//		$body
//	}
//
// =>
//
//	for it := $X ; it.Next(); {
//		x [:]= it.Current()
//		$body
//	}
func (r *rewriter) rewriteForRange(pkg loader.Pkg, fr *ast.RangeStmt) *ast.ForStmt {
	isValid := fr.Key != nil && fr.Value == nil
	r.assert(pkg, isValid, fr, "invalid for range")

	// iter := X.Ident(cstIterVar)
	iter := pkg.NewIdent(cstIterVar, pkg.TypeOf(fr.X))
	current := X.Select(iter, cstCurrent)
	next := X.Select(iter, cstMoveNext)

	init := X.Define(iter, fr.X)
	cond := X.Call(next)
	body := X.Block1(
		X.Assign(fr.Tok, fr.Key, X.Call(current)),
		fr.Body.List...,
	)
	return X.ForStmt(init, cond, nil, body)
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewrite co.Iter ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// co.Iter[T] => seq.Iterator[T]
func (r *rewriter) rewriteIter(c *astutil.Cursor, pkg loader.Pkg) bool {
	switch n := c.Node().(type) {
	case *ast.IndexExpr:
		if r.isIterator(pkg.TypeOf(n.X)) {
			c.Replace(X.Index(
				X.PkgSelect(r.seqImportedName, cstIterator),
				n.Index,
			))
		}
		return true
	}
	return true
}
