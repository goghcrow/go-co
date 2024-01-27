package rewriter

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goghcrow/go-ast-matcher"
	"github.com/goghcrow/go-ast-matcher/imports"
	"golang.org/x/tools/go/ast/astutil"
)

func Compile(
	inputDir, outputDir string,
	patterns []string,
	opts ...matcher.MatchOption,
) {
	flags := log.Flags()
	defer log.SetFlags(flags)
	log.SetFlags(0)

	tmpOut := outputDir + "_tmp"
	_ = os.RemoveAll(tmpOut)

	log.SetPrefix("[rewrite] ")
	rewrite(inputDir, tmpOut, patterns, opts...)

	// type info broken after rewriting, so reload to optimize
	log.SetPrefix("[optimize] ")
	optimize(tmpOut, outputDir, patterns, opts...)
}

func newMatcher(inputDir, outputDir string,
	patterns []string,
	opts ...matcher.MatchOption,
) *matcher.Matcher {
	inputDir, err := filepath.Abs(inputDir)
	panicIf(err)

	outputDir, err = filepath.Abs(outputDir)
	panicIf(err)

	if inputDir == outputDir {
		panic("overwrite")
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	panicIf(err)

	return matcher.NewMatcher(
		inputDir,
		patterns,
		append(opts, matcher.WithLoadDepts())...,
	)
}

func rewrite(
	inputDir, outputDir string,
	patterns []string,
	opts ...matcher.MatchOption,
) {
	m := newMatcher(inputDir, outputDir, patterns, opts...)

	r := &rewriter{
		Matcher:       m,
		iterType:      m.MustLookup(qualifiedIter),
		yieldFunc:     m.MustLookup(qualifiedYield),
		yieldFromFunc: m.MustLookup(qualifiedYieldFrom),
	}

	coPkg := r.All[pkgCoPath]
	if coPkg == nil {
		log.Printf("skip rewrite: no import %s\n", pkgCoPath)
		return
	}

	parseOrImport := func(f *ast.File) (coName, seqName string) {
		coName = imports.ImportName(f, pkgCoPath, pkgCoName)
		assert(coName != "") // coPkg != nil
		seqName = imports.ImportName(f, pkgSeqPath, pkgSeqName)
		if seqName == "" {
			seqName = importSeqName
			astutil.AddNamedImport(m.FSet, f, importSeqName, pkgSeqPath)
		}
		return
	}

	r.VisitAllFiles(func(m *matcher.Matcher, file *ast.File) {
		if !imports.Uses(m, file, coPkg.Types) {
			log.Printf("skip file: %s\n", r.Filename)
			return
		}

		// 1. init file context
		r.coImportedName, r.seqImportedName = parseOrImport(file) // parse import name
		r.comments = nil

		r.yieldFuncDecls = map[*ast.FuncDecl]bool{}
		r.yieldFuncLits = map[*ast.FuncLit]bool{}
		r.collectYieldFunc(file) // collect func with yield/yieldFrom call

		// 2. rewrite file
		log.Printf("visit file: %s\n", r.Filename)
		do := func(f astutil.ApplyFunc) { astutil.Apply(file, nil, f) }

		// file level instance for file scope cache
		// notice: order matters
		do(r.attachComment)        // attach the original source to comments
		do(mkYieldFromRewriter(r)) // rewrite yieldFrom() to range yield() (range co.Iter)
		do(r.rewriteForRanges)     // rewrite range co.Iter to for loop co.Iter
		do(mkYieldRewriter(r))     // rewrite yield func
		do(r.rewriteIter)          // rewrite all co.Iter to seq.Iterator

		// 3. write file
		log.Printf("write file: %s\n", r.Filename)
		// clear free-floating comments, preventing confusing position of comments
		// https://github.com/golang/go/issues/20744
		r.File.Comments = r.comments
		filename := strings.ReplaceAll(r.Filename, m.Cfg.Dir, outputDir)
		r.WriteGeneratedFile(filename, pkgCoPath)
	})
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewriter ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

type rewriter struct {
	*matcher.Matcher

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

// Yield is stmt, not expr
func (r *rewriter) isYieldCall(n ast.Node) (*ast.CallExpr, bool) {
	// e.g. for Yield(1); not here; Yield(2) {  Yield(3) }
	return isCallStmtOf(r.Info, n, r.yieldFunc)
}

// YieldFrom is stmt, not expr
func (r *rewriter) isYieldFromCall(n ast.Node) (*ast.CallExpr, bool) {
	return isCallStmtOf(r.Info, n, r.yieldFromFunc)
}

func (r *rewriter) isYieldFuncDecl(f *ast.FuncDecl) bool {
	return r.yieldFuncDecls[f]
}

func (r *rewriter) isYieldFuncLit(f *ast.FuncLit) bool {
	return r.yieldFuncLits[f]
}

func (r *rewriter) yieldFuncRetParamTy(f *ast.FuncType) ast.Expr {
	retTy := f.Results.List[0].Type
	idx, is := retTy.(*ast.IndexExpr)
	r.assert(is, f, "invalid yield func type")
	return idx.Index
}

func (r *rewriter) isIterator(ty types.Type) bool {
	return identicalWithoutTypeParam(r.iterType.Type(), ty)
}

func (r *rewriter) assert(ok bool, pos any, format string, a ...any) {
	if !ok {
		loc := "unknown"
		if !isNil(pos) {
			switch pos := pos.(type) {
			case ast.Node:
				loc = r.ShowPos(pos)
			case token.Pos:
				loc = r.FSet.Position(pos).String()
			case string:
				loc = pos
			}
		}
		panic(fmt.Sprintf(format, a...) + " in: " + loc)
	}
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Collect YieldFunc ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// collect funs which containing yield / yieldFrom called directly by reference,
// make sure funcDecl | funcLit modified in place afterward
func (r *rewriter) collectYieldFunc(file *ast.File) {
	var (
		yieldFunStack = mkStack[ast.Node /*FuncDecl|FuncLit*/](nil)
		enter         = yieldFunStack.push
		exit          = yieldFunStack.pop
		outer         = yieldFunStack.top
	)

	checkSignature := func(funTy types.Type, pos token.Pos) {
		msg := "invalid yield func signature, expect one co.Iter[T] return"

		sig, ok := funTy.(*types.Signature)
		r.assert(ok, pos, msg)
		rs := sig.Results()

		singleRet := rs != nil && rs.Len() == 1
		r.assert(singleRet, pos, msg)

		retTy := rs.At(0).Type()
		retIter := r.isIterator(retTy)
		r.assert(retIter, pos, msg)
	}

	cache := map[ast.Node]bool{}
	astutil.Apply(file, func(c *astutil.Cursor) bool {
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
			callee := r.ObjectOfCall(n)
			if callee == r.yieldFunc || callee == r.yieldFromFunc {
				switch f := outer().(type) {
				case *ast.FuncDecl:
					checkSignature(r.TypeOf(f.Name), n.Pos())
					r.yieldFuncDecls[f] = true
				case *ast.FuncLit:
					checkSignature(r.TypeOf(f), n.Pos())
					r.yieldFuncLits[f] = true
				}
			}
		}
		return true
	})
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Attach comment ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func (r *rewriter) attachComment(c *astutil.Cursor) bool {
	if runningWithGoTest {
		return true
	}
	switch f := c.Node().(type) {
	case *ast.FuncDecl:
		if r.isYieldFuncDecl(f) {
			// attach comment to func decl
			src := r.ShowNode(f)
			X.AppendComment(&f.Doc, X.Comment(f.Pos()-1, src))
			c.Replace(f)
		}
		return true
	case *ast.FuncLit:
		if r.isYieldFuncLit(f) {
			// attach comment to free-float
			src := r.ShowNode(f)
			r.comments = append(r.comments,
				X.Comments(X.Comment(f.Pos()-1, src)))
			c.Replace(f)
		}
		return true
	}
	return true
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewrite ForInit ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

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
func (r *rewriter) rewriteInitStmt(c *astutil.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ForStmt:
		if isDefineStmt(n.Init) {
			init := n.Init
			n.Init = nil
			n.For = token.NoPos
			c.Replace(X.Block(init, n))
		}
	case *ast.SwitchStmt:
		if isDefineStmt(n.Init) {
			init := n.Init
			n.Init = nil
			n.Switch = token.NoPos
			c.Replace(X.Block(init, n))
		}
	case *ast.TypeSwitchStmt:
		if isDefineStmt(n.Init) {
			init := n.Init
			n.Init = nil
			n.Switch = token.NoPos
			c.Replace(X.Block(init, n))
		}
	}
	return true
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewrite co.Iter ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// co.Iter[T] => seq.Iterator[T]
func (r *rewriter) rewriteIter(c *astutil.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.IndexExpr:
		if r.isIterator(r.TypeOf(n.X)) {
			c.Replace(X.Index(
				X.PkgSelect(r.seqImportedName, cstIterator),
				n.Index,
			))
		}
		return true
	}
	return true
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Rewrite Range Generator ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func (r *rewriter) rewriteForRanges(c *astutil.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.RangeStmt:
		if r.isIterator(r.TypeOf(n.X)) {
			c.Replace(r.rewriteForRange(n))
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
func (r *rewriter) rewriteForRange(fr *ast.RangeStmt) *ast.ForStmt {
	isValid := fr.Key != nil && fr.Value == nil
	r.assert(isValid, fr, "invalid for range")

	// iter := X.Ident(cstIterVar)
	iter := r.NewIdent(cstIterVar, r.TypeOf(fr.X))
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
