package rewriter

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/goghcrow/go-ast-matcher"
)

func TestRewrite(t *testing.T) {
	in := "test/src"
	out := "test/out"
	tmp := out + "_tmp"
	files := matcher.PatternAll

	Compile(in, out, files, matcher.WithLoadTest())

	xs, _ := os.ReadDir(in)
	for _, x := range xs {
		if x.IsDir() {
			continue
		}

		if strings.HasSuffix(x.Name(), ".go.tmp") {
			expect, _ := os.ReadFile(path.Join(in, x.Name()))
			output, err := os.ReadFile(path.Join(tmp, strings.Split(x.Name(), ".")[0]+".go"))
			if err != nil {
				panic(err)
			}
			if string(output) != string(expect) {
				t.Fatalf(x.Name())
			}
		}

		if strings.HasSuffix(x.Name(), ".go.out") {
			expect, _ := os.ReadFile(path.Join(in, x.Name()))
			output, err := os.ReadFile(path.Join(out, strings.Split(x.Name(), ".")[0]+".go"))
			if err != nil {
				panic(err)
			}
			if string(output) != string(expect) {
				t.Fatalf(x.Name())
			}
		}
	}
}
