package rewriter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goghcrow/go-ast-matcher"
	"github.com/goghcrow/go-loader"
	"github.com/goghcrow/go-matcher"
)

const fileComment = `//go:build !%s

// Code generated by github.com/goghcrow/go-co DO NOT EDIT.
`

type FilePrinter func(filename string, f *loader.File)

func Compile(srcDir, dstDir string, opts ...loader.Option) {
	srcDir, err := filepath.Abs(srcDir)
	panicIf(err)

	dstDir = mustMkDir(dstDir)

	tmpOutputDir := mustMkDir(dstDir + "_tmp")
	// tmpOutputDir, err = os.MkdirTemp("", "co_")
	// panicIf(err)
	if !runningWithGoTest {
		//goland:noinspection GoUnhandledErrorResult
		defer os.RemoveAll(tmpOutputDir)
	}

	resetLog()
	log.SetPrefix("[rewrite] ")
	r := mkRewriter(astmatcher.New(
		loader.MustNew(srcDir, append(opts, loader.WithLoadDepts())...),
		matcher.New(),
	))

	comment := fmt.Sprintf(fileComment, defaultBuildTag)
	r.rewriteAllFiles(func(filename string, f *loader.File) {
		filename = strings.ReplaceAll(filename, srcDir, tmpOutputDir)
		f.WriteWithComment(filename, comment)
	})

	// type info broken after rewriting, so reload to optimize
	log.SetPrefix("[optimize] ")
	o := mkOptimizer(astmatcher.New(
		loader.MustNew(tmpOutputDir, append(opts, loader.WithLoadDepts(), loader.WithSuppressErrors())...),
		matcher.New(),
	))
	o.optimizeAllFiles(func(filename string, f *loader.File) {
		filename = strings.ReplaceAll(filename, tmpOutputDir, dstDir)
		f.WriteWithComment(filename, comment)
	})
}

type (
	Option func(*option)
	option struct {
		fileSuffix string
		buildTag   string
	}
)

func WithFileSuffix(s string) Option { return func(opt *option) { opt.fileSuffix = s } }
func WithBuildTag(s string) Option   { return func(opt *option) { opt.buildTag = s } }

const (
	defaultFileSuffix = "co"
	defaultBuildTag   = "co"
)

func GoGen(dir string, opts ...Option) {
	opt := &option{
		fileSuffix: defaultFileSuffix,
		buildTag:   defaultBuildTag,
	}
	for _, o := range opts {
		o(opt)
	}

	tmpOutputDir := mustMkDir(dir + "_tmp")
	if !runningWithGoTest {
		//goland:noinspection GoUnhandledErrorResult
		defer os.RemoveAll(tmpOutputDir)
	}

	var (
		endsWith       = strings.HasSuffix
		replace        = strings.ReplaceAll
		comment        = fmt.Sprintf(fileComment, opt.buildTag)
		srcFileSuffix  = fmt.Sprintf("_%s.go", opt.fileSuffix)
		testFileSuffix = fmt.Sprintf("_%s_test.go", opt.fileSuffix)
		isCoFile       = func(filename string) bool {
			return endsWith(filename, srcFileSuffix) || endsWith(filename, testFileSuffix)
		}
	)

	resetLog()
	log.SetPrefix("[rewrite] ")
	r := mkRewriter(astmatcher.New(
		loader.MustNew(dir,
			loader.WithLoadDepts(),
			loader.WithLoadTest(),
			loader.WithBuildTag(opt.buildTag),
			loader.WithFileFilter(func(f *loader.File) bool { return isCoFile(f.Filename) }),
		),
		matcher.New(),
	))
	r.rewriteAllFiles(func(filename string, f *loader.File) {
		filename = replace(filename, srcFileSuffix, ".go")
		filename = replace(filename, testFileSuffix, "_test.go")
		filename = replace(filename, dir, tmpOutputDir)
		f.WriteWithComment(filename, comment)
	})

	log.SetPrefix("[optimize] ")
	o := mkOptimizer(astmatcher.New(
		loader.MustNew(tmpOutputDir,
			loader.WithLoadDepts(),
			loader.WithLoadTest(),
			loader.WithSuppressErrors(),
		),
		matcher.New(),
	))
	o.optimizeAllFiles(func(filename string, f *loader.File) {
		filename = strings.ReplaceAll(filename, tmpOutputDir, dir)
		f.WriteWithComment(filename, comment)
	})
}

func resetLog() {
	flags := log.Flags()
	defer log.SetFlags(flags)
	log.SetFlags(0)
}
