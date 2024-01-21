package rewriter

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"reflect"
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

		// 1. parse import name
		r.coImportedName, r.seqImportedName = parseOrImport(file)

		// file level instance for file scope cache
		rewriteYield := mkYieldRewriter(r)
		rewriteYieldFrom := mkYieldFromRewriter(r)

		// 2. rewrite file
		log.Printf("visit file: %s\n", r.Filename)
		do := func(f astutil.ApplyFunc) { astutil.Apply(file, nil, f) }
		// notice: order matters
		do(rewriteYieldFrom.rewrite) // rewrite yieldFrom() to range yield() (range co.Iter)
		do(r.rewriteForRanges)       // rewrite range co.Iter to for loop co.Iter
		do(r.rewriteInitStmt)        // extract out define in for-init/switch-init
		do(rewriteYield.rewrite)     // rewrite yield func
		do(r.rewriteIter)            // rewrite all co.Iter to seq.Iterator

		// 3. write file
		log.Printf("write file: %s\n", r.Filename)
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

// isGenerator
func (r *rewriter) isYieldFunc(funTy types.Type) bool {
	sig, ok := funTy.(*types.Signature)
	if !ok {
		return false
	}

	rs := sig.Results()

	singleRet := rs != nil && rs.Len() == 1
	if !singleRet {
		return false
	}

	retTy := rs.At(0).Type()
	return r.isIterator(retTy)
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

func (r *rewriter) updateUses(idOrSel ast.Expr, obj types.Object) {
	switch x := idOrSel.(type) {
	case *ast.Ident:
		r.Uses[x] = obj
	case *ast.SelectorExpr:
		r.Uses[x.Sel] = obj
	default:
		panic("unreached")
	}
}

func (r *rewriter) assert(ok bool, pos ast.Node, format string, a ...any) {
	if !ok {
		loc := "unknown"
		if pos != nil && reflect.ValueOf(pos).IsNil() {
			loc = r.ShowPos(pos)
		}
		panic(fmt.Sprintf(format, a...) + " in: " + loc)
	}
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

	iter := X.Ident(cstIterVar)
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