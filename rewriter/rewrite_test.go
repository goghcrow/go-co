package rewriter

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/goghcrow/go-ast-matcher"
)

func TestTest(t *testing.T) {
	Compile("./test/src",
		"./test/out",
		matcher.PatternAll,
		matcher.WithLoadTest(),
	)
}

func TestRewrite(t *testing.T) {
	in, out := "test/src", "test/out"
	files := matcher.PatternAll

	Compile(in, out, files, matcher.WithLoadTest())

	xs, _ := os.ReadDir(in)
	for _, x := range xs {
		if x.IsDir() {
			continue
		}
		if strings.HasSuffix(x.Name(), ".expect") {
			expect, _ := os.ReadFile(path.Join(in, x.Name()))
			output, err := os.ReadFile(path.Join(out+"_tmp", strings.Split(x.Name(), ".")[0]+".go"))
			if err != nil {
				panic(err)
			}
			if string(output) != string(expect) {
				t.Fatalf(x.Name())
			}
		}
	}
}