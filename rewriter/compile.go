package rewriter

import (
	"fmt"
	"go/ast"
	"log"
	"os"
	"path/filepath"
	"strings"

	matcher "github.com/goghcrow/go-ast-matcher"
)

const fileComment = `//go:build !co

// Code generated by github.com/goghcrow/go-co DO NOT EDIT.
`

const (
	fileSuffix = "co"
	buildTag   = "co"
)

var (
	srcFileSuffix  = fmt.Sprintf("_%s.go", fileSuffix)
	testFileSuffix = fmt.Sprintf("_%s_test.go", fileSuffix)
)

type FilePrinter func(filename string, file *ast.File)

func Compile(
	srcDir, dstDir string,
	patterns []string,
	opts ...matcher.MatchOption,
) {
	resetLog()

	srcDir, err := filepath.Abs(srcDir)
	panicIf(err)

	dstDir = mkDir(dstDir)

	tmpOutputDir := mkDir(dstDir + "_tmp")
	// tmpOutputDir, err = os.MkdirTemp("", "co_")
	// panicIf(err)
	if !runningWithGoTest {
		defer os.RemoveAll(tmpOutputDir)
	}

	log.SetPrefix("[rewrite] ")
	r := mkRewriter(matcher.NewMatcher(
		srcDir,
		patterns,
		append(opts, matcher.WithLoadDepts())...,
	))
	r.rewriteAllFiles(func(filename string, file *ast.File) {
		filename = strings.ReplaceAll(filename, srcDir, tmpOutputDir)
		r.WriteFileWithComment(filename, fileComment)
	})

	// type info broken after rewriting, so reload to optimize
	log.SetPrefix("[optimize] ")
	o := mkOptimizer(matcher.NewMatcher(
		tmpOutputDir,
		patterns,
		append(opts, matcher.WithSuppressErrors(), matcher.WithLoadDepts())...,
	))
	o.optimizeAllFiles(func(filename string, file *ast.File) {
		filename = strings.ReplaceAll(filename, tmpOutputDir, dstDir)
		o.WriteFileWithComment(filename, fileComment)
	})
}

func GoGen(dir string) {
	resetLog()

	tmpOutputDir := mkDir(dir + "_tmp")
	if !runningWithGoTest {
		//goland:noinspection GoUnhandledErrorResult
		defer os.RemoveAll(tmpOutputDir)
	}

	log.SetPrefix("[rewrite] ")
	r := mkRewriter(matcher.NewMatcher(
		dir,
		matcher.PatternAll,
		matcher.WithLoadDepts(),
		matcher.WithLoadTest(),
		matcher.WithBuildTag(buildTag),
		matcher.WithFileFilter(func(filename string, file *ast.File) bool {
			return strings.HasSuffix(filename, srcFileSuffix) ||
				strings.HasSuffix(filename, testFileSuffix)
		}),
	))
	r.rewriteAllFiles(func(filename string, file *ast.File) {
		filename = strings.ReplaceAll(filename, srcFileSuffix, ".go")
		filename = strings.ReplaceAll(filename, testFileSuffix, "_test.go")
		filename = strings.ReplaceAll(filename, dir, tmpOutputDir)
		r.WriteFileWithComment(filename, fileComment)
	})

	log.SetPrefix("[optimize] ")
	o := mkOptimizer(matcher.NewMatcher(
		tmpOutputDir,
		matcher.PatternAll,
		matcher.WithLoadDepts(),
		matcher.WithLoadTest(),
		matcher.WithSuppressErrors(),
	))
	o.optimizeAllFiles(func(filename string, file *ast.File) {
		filename = strings.ReplaceAll(filename, tmpOutputDir, dir)
		o.WriteFileWithComment(filename, fileComment)
	})
}

func resetLog() {
	flags := log.Flags()
	defer log.SetFlags(flags)
	log.SetFlags(0)
}
